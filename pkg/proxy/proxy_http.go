package proxy

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/http2"

	uuid "github.com/nu7hatch/gouuid"

	"github.com/schubergphilis/mercury/pkg/healthcheck"
	"github.com/schubergphilis/mercury/pkg/logging"
)

const (
	sessionIDCookie = "mercid"
)

type customTransport struct {
	*http.Transport
	LocalAddr net.Addr
}

var variableRegex = regexp.MustCompile("###([A-Z_a-z]+)###")

func customStatusPage(statusCode int, statusMessage string, req *http.Request) *http.Response {
	var body []byte
	nbody := &bytes.Buffer{}
	t := time.Now()
	msg := fmt.Sprintf("<head><title>%d %s</title></head><body><h1>%d %s</h1><br>- Generated by Mercury at %s</body>", statusCode, statusMessage, statusCode, statusMessage, t.Format("2006-01-02 15:04:05"))
	nbody.Write(append(body, []byte(msg)...))
	b := ioutil.NopCloser(nbody)
	nres := &http.Response{
		StatusCode: statusCode,
		Status:     statusMessage,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     http.Header{},
		Body:       b,
		Request:    req,
	}
	// Ensure we dont cache custom responses
	nres.Header.Add("Cache-Control", "no-cache, no-store, must-revalidate")
	nres.Header.Add("Pragma", "no-cache")
	nres.Header.Add("Expires", "0")

	return nres

}

// RoundTrip does the actual http sending and receiving for the proxy
func (t *customTransport) RoundTrip(req *http.Request) (res *http.Response, err error) {
	remoteAddr := strings.Split(req.RemoteAddr, ":")
	log := logging.For("proxy/roundtrip").WithField("clientip", remoteAddr[0])
	log.WithField("scheme", req.URL).Debug("Roundtrip scheme")
	starttime := time.Now()
	originalScheme := req.URL.Scheme
	// Process the scheme and see if its an error, internal or other
	// Errors are generated if no backend could be found
	// Internal is done for ACL-only connections (such as redirect)
	// Default is all other http connections to a remote host
	scheme := strings.Split(req.URL.Scheme, "//")
	switch scheme[0] {
	case "error":
		log.WithField("errorcode", scheme[1]).WithField("error", scheme[2]).Infof("Could not proxy client request")
		var statuscode int
		statusmessage := scheme[3]
		statuscode, err = strconv.Atoi(scheme[2])
		if err != nil {
			statuscode = 502
			statusmessage = http.StatusText(statuscode)
		}
		res = customStatusPage(statuscode, statusmessage, req)
		return res, nil

	case "maintenance":
		log.WithField("errorcode", scheme[1]).WithField("error", scheme[2]).Infof("Could not proxy client request")
		var statuscode int
		statusmessage := scheme[3]
		statuscode = 503
		res = customStatusPage(statuscode, statusmessage, req)
		return res, nil

	case "internal":
		res = customStatusPage(200, "OK", req)

	default: // http/https
		req.URL.Scheme = scheme[0]
		res, err = t.Transport.RoundTrip(req)
		if err != nil {
			// We have an error, generate a 500
			res = customStatusPage(500, err.Error(), req)
		}

		log = log.WithField("scheme", req.URL.Scheme)
	}
	// At this point res can never by nil

	// Add clientid (mercid) cookie to logging
	if clientid, cerr := req.Cookie(sessionIDCookie); cerr == nil {
		log = log.WithField("clientid", clientid.Value)
	}

	// Log request
	roundtriptime := time.Since(starttime)
	log = log.WithField("backendnode", req.URL.Hostname()).WithField("forwarded-for", req.Header.Get("X-Forwarded-for")).WithField("hostname", req.Host).WithField("method", req.Method).WithField("url", req.RequestURI)
	log = log.WithField("clientproto", req.Proto).WithField("referer", req.Referer()).WithField("useragent", req.Header.Get("User-Agent"))
	log = log.WithField("statuscode", res.StatusCode).WithField("contentlength", res.ContentLength).WithField("serverproto", res.Proto)
	log.WithField("roundtriptime", roundtriptime.Seconds()).Info("HTTP response")

	// Save the original scheme, we need it when modifying output
	res.Request.URL.Scheme = originalScheme
	if res.Request.Header == nil {
		res.Request.Header = req.Header
	}
	return res, err
}

// processACLVariables converts ###TAG###'s in to values based on backendnode
func processACLVariables(acl []ACL, l *Listener, backendnode BackendNode, req *http.Request) []ACL {
	log := logging.For("proxy/aclvariables").WithField("pool", l.Name).WithField("localip", l.IP).WithField("localport", l.Port).WithField("mode", l.ListenerMode)

	fn := func(m string) string {
		name := variableRegex.FindStringSubmatch(m)[1]
		value, err := getVariableValue(name, l, &backendnode, req)
		if err != nil {
			log.WithField("variable", name).WithField("error", err).Warn("Unable to get variable value")
			return name
		}
		return value
	}

	var newACL []ACL

	for _, acl := range acl {
		// regex conversion
		// header value
		if acl.HeaderValue != "" {
			newdata := variableRegex.ReplaceAllStringFunc(acl.HeaderValue, fn)
			acl.HeaderValue = newdata
		}
		// cookie value
		if acl.CookieValue != "" {
			newdata := variableRegex.ReplaceAllStringFunc(acl.CookieValue, fn)
			acl.CookieValue = newdata
		}
		// append new line to acl
		newACL = append(newACL, acl)
	}

	return newACL
}

func getVariableValue(name string, l *Listener, backendnode *BackendNode, req *http.Request) (string, error) {
	switch name {
	case "NODE_ID":
		return backendnode.UUID, nil

	case "NODE_IP":
		return backendnode.IP, nil

	case "LB_IP":
		return l.IP, nil

	case "LB_PORT":
		return strconv.Itoa(l.Port), nil

	case "REQ_URL":
		return req.Host + req.URL.Path, nil

	case "REQ_QUERY":
		return req.URL.RawQuery, nil

	case "REQ_PATH":
		return req.URL.Path, nil

	case "REQ_HOST":
		val := strings.Split(req.Host, ":")
		return val[0], nil

	case "REQ_IP":
		val := strings.Split(req.Host, ":")
		return val[0], nil

	case "REQ_PROTO":
		if req.TLS != nil {
			return "https", nil
		}
		return "http", nil

	case "CLIENT_IP":
		val := strings.Split(req.RemoteAddr, ":")
		return val[0], nil

	case "CLIENT_CERT":
		return getClientCertValue(req)

	case "UUID":
		id, uerr := uuid.NewV4() // used for sticky cookies
		if uerr == nil {
			return id.String(), nil
		}
		return "", uerr
	}

	// return same if no correct match
	return "", fmt.Errorf("Unknown variable: %v", name)
}

func getClientCertValue(req *http.Request) (string, error) {
	if req.TLS == nil {
		return "", nil
	}

	var values []string

	for _, cert := range req.TLS.PeerCertificates {
		certPEM, err := getCertificatePEM(cert)
		if err != nil {
			return "", err
		}
		values = append(values, certPEM)
	}

	return strings.Join(values, ","), nil
}

func getCertificatePEM(cert *x509.Certificate) (string, error) {
	block := pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw}

	certPEM := pem.EncodeToMemory(&block)
	if certPEM == nil {
		return "", errors.New("Cannot extract certificate content")
	}

	replacer := strings.NewReplacer("-----BEGIN CERTIFICATE-----", "",
		"-----END CERTIFICATE-----", "",
		"\n", "")

	cleanPEM := replacer.Replace(string(certPEM))

	return cleanPEM, nil
}

func addClientSessionID(req *http.Request, res *http.Response, id string) {
	if res == nil {
		return
	}

	expire := time.Now().Add(24 * time.Hour)
	sessionCookie := &http.Cookie{
		Name:     sessionIDCookie,
		Value:    id,
		Path:     "/",
		Expires:  expire,
		HttpOnly: true}
	if strings.EqualFold(req.URL.Scheme, "https") {
		sessionCookie.Secure = true
	}

	res.Header.Add("Set-Cookie", sessionCookie.String())
}

// NewHTTPProxy Create a HTTP proxy
func (l *Listener) NewHTTPProxy() *httputil.ReverseProxy {
	log := logging.For("proxy/httpproxy").WithField("pool", l.Name)

	// directory is the main handler,
	// it :
	// - finds the backend for the client based on the requested
	// - applies a client Cookie
	// - applies inbound ACL's (to be sent to the backend server)
	// - sets the url Scheme to be processed by the RoundTrip handler, and the ModifyResponse handler
	director := func(req *http.Request) {
		remoteAddr := strings.Split(req.RemoteAddr, ":")
		clog := log.WithField("clientip", remoteAddr[0]).WithField("hostname", req.Host)
		// Update statistics of the Listener
		l.Statistics.ClientsConnectsAdd(1)
		l.updateClients()

		// Log clients request
		clientid, err := req.Cookie(sessionIDCookie)
		if err == nil {
			clog = clog.WithField("clientid", clientid.Value)
		} else {
			clog = clog.WithField("clientid", "")
		}

		rlog := clog.WithField("hostname", req.URL.Hostname()).WithField("forwarded-for", req.Header.Get("X-Forwarded-for")).WithField("method", req.Method).WithField("url", req.RequestURI)
		rlog = rlog.WithField("proto", req.Proto).WithField("contentlength", req.ContentLength).WithField("referer", req.Referer()).WithField("useragent", req.Header.Get("User-Agent"))
		rlog.Info("HTTP request")

		// Check if we have a host
		if req.Host == "" {
			clog.Warn("Request done without host header, disconnecting client")
			req.URL.Scheme = "error//unknown//400//Invalid request - no host was supplied"
			return
		}

		// we have a host, find it's matching backend
		reqHost := strings.Split(req.Host, ":")
		backendname, backend := l.FindBackendByHost(reqHost[0])
		clog = clog.WithField("backend", backendname)
		if backendname == "" {
			// We don't have a backend match, this could be due to a hostname in the request which is unknown, and only if there is no default
			other := l.FindAllHostNames()
			clog.WithField("requested", reqHost[0]).WithField("available", strings.Join(other, ", ")).Error("Unable to find backend for requested host")
			req.URL.Scheme = "error//" + backendname + "//503//Service Unavailable - no backend found"
			return
		}

		var stickyCookie string
		if strings.Contains(backend.BalanceMode, "sticky") {
			// Check for the stky cookie, used for sticky session, only if we have sticky loadbalancing
			stky, serr := req.Cookie("stky")
			if serr == nil {
				stickyCookie = stky.Value
			}
		}

		// Internal requests TODO: ease up this use
		if backend.ConnectMode == "internal" {
			clog.Debug("Internal request")
			req.URL.Scheme = fmt.Sprintf("internal//%s//%s", backendname, req.URL.Hostname())
			return
		}

		// Get a Node to balance this request to
		client := strings.Split(req.RemoteAddr, ":")
		backendnode, status, err := backend.GetBackendNodeBalanced(backendname, client[0], stickyCookie, backend.BalanceMode)
		if err != nil {
			clog.WithField("error", err).Error("No backend node available")
			if status == healthcheck.Maintenance {
				req.URL.Scheme = "maintenance//" + backendname + "//503//Service Unavailable - no backend available"
				return
			}
			req.URL.Scheme = "error//" + backendname + "//503//Service Unavailable - no backend available"
			return
		}
		clog.WithField("backendip", backendnode.IP).WithField("backendport", backendnode.Port).Debug("Forwarding HTTP request to backend")

		acl := processACLVariables(l.Backends[backendname].InboundACL, l, *backendnode, req)
		aclAllows := l.Backends[backendname].InboundACL.CountActions("allow")
		aclDenies := l.Backends[backendname].InboundACL.CountActions("deny")

		// Process all ACL's and count hit's if any
		aclsHit := 0
		for _, inacl := range acl {
			if inacl.ProcessRequest(req) { // process request returns true if we match a allow/deny acl
				aclsHit++
			}
		}

		// Take actions based on allow/deny, you cannot combine allow and denies
		if aclDenies > 0 && aclAllows > 0 {
			log.Errorf("Found ALLOW and DENY ACL's in the same block, only allows will be processed")
		}

		if aclAllows > 0 && aclsHit == 0 { // setting an allow ACL, will deny all who do not match atleast 1 allow
			req.URL.Scheme = "error//" + backendname + "//403//Access denied - does not match ALLOW ACL"
			clog.Infof("Client did not match allow acl")
			return
		} else if aclAllows == 0 && aclDenies > 0 && aclsHit > 0 { // setting an deny ACL, will deny all who match 1 of the denies
			clog.Infof("Client matched deny acl")
			req.URL.Scheme = "error//" + backendname + "403//Access denied - matched DENY ACL"
			return
		}

		backendnode.Statistics.ClientsConnectsAdd(1)
		backendnode.Statistics.TimeCounterAdd() // connections past 30 seconds
		backendnode.Statistics.ClientsConnectedAdd(1)
		reqDump, err := httputil.DumpRequest(req, true)
		if err != nil {
			backendnode.Statistics.RXAdd(int64(len(reqDump)))
		}

		clog.WithField("statistics", fmt.Sprintf("%+v", backendnode.Statistics)).Debug("Statistics updated")
		req.URL.Scheme = fmt.Sprintf("%s//%s//%s", backend.ConnectMode, backendname, backendnode.UUID)
		req.URL.Host = fmt.Sprintf("%s:%d", backendnode.IP, backendnode.Port)
	}

	modifyresponse := func(res *http.Response) error {
		// Process OutboundACL if we have a valid request (does not apply to errors)
		localerror := false
		localmaintenance := false
		var errorpage []byte
		var maintenancepage []byte
		var showerrorpage bool
		if res.Request != nil {
			scheme := strings.Split(res.Request.URL.Scheme, "//")
			proto := scheme[0]
			backendname := scheme[1]
			nodeid := scheme[2]

			// Check if backend has error/maintenance page, if so, keep it
			if _, ok := l.Backends[backendname]; ok {
				if l.Backends[backendname].ErrorPage.present() {
					errorpage = l.Backends[backendname].ErrorPage.content
					showerrorpage = l.Backends[backendname].ErrorPage.threshold(res.StatusCode)
				}
				if l.Backends[backendname].MaintenancePage.present() {
					maintenancepage = l.Backends[backendname].MaintenancePage.content
				}
			}

			switch proto {
			case "maintenance":
				localmaintenance = true
			case "error":
				localerror = true
			default:
				if backendname != "localhost" && backendname != "" {
					// Get ACL's
					acls := l.Backends[backendname].OutboundACL
					node, err := l.Backends[backendname].GetBackendNodeByID(nodeid)
					if err != nil {
						log.WithError(err).Debug("Did not parse node ACL, since no node could be found:")
						node = &BackendNode{}
					}
					// Change ACL's to processed variables
					acls = processACLVariables(acls, l, *node, res.Request)
					// Apply ACL
					for _, acl := range acls {
						acl.ProcessResponse(res)
					}
				}

			}
		}

		// Load listener error page if any, and no backend error page
		if len(errorpage) == 0 {
			if l.ErrorPage.present() {
				errorpage = l.ErrorPage.content
				showerrorpage = l.ErrorPage.threshold(res.StatusCode)
			}
		}

		// Load listener maintenance page if any, and no backend maintenance page
		if len(maintenancepage) == 0 {
			if l.MaintenancePage.present() {
				maintenancepage = l.MaintenancePage.content
			}
		}

		if localmaintenance {
			if len(maintenancepage) > 0 { // show maintenance page
				nbody := &bytes.Buffer{}
				nbody.Write(maintenancepage)
				res.Header.Add("x-statuscode", fmt.Sprintf("%d", res.StatusCode))
				res.Header.Add("x-statusmessage", res.Status)
				res.Body = ioutil.NopCloser(nbody)
				// force content length to new size of error body
				res.Header.Set("Content-Length", fmt.Sprintf("%d", len(maintenancepage)))
				res.Header.Add("Cache-Control", "no-cache, no-store, must-revalidate")
				res.Header.Add("Pragma", "no-cache")
				res.Header.Add("Expires", "0")
				return nil
			}
			// No maintenance page, default back to error page instead
			localerror = true
		}

		// Alternative ErrorPage if statuscode reached threshold or local error
		if len(errorpage) > 0 && (showerrorpage || localerror == true) {
			nbody := &bytes.Buffer{}
			nbody.Write(errorpage)
			res.Header.Add("x-statuscode", fmt.Sprintf("%d", res.StatusCode))
			res.Header.Add("x-statusmessage", res.Status)
			res.Body = ioutil.NopCloser(nbody)
			// force content length to new size of error body
			res.Header.Set("Content-Length", fmt.Sprintf("%d", len(errorpage)))
			res.Header.Add("Cache-Control", "no-cache, no-store, must-revalidate")
			res.Header.Add("Pragma", "no-cache")
			res.Header.Add("Expires", "0")
		}

		return nil
	}

	proxy := func(req *http.Request) (*url.URL, error) {
		return http.ProxyFromEnvironment(req)
	}

	tlsClientConfig := &tls.Config{
		InsecureSkipVerify: true,
	}

	var localAddr *net.IPAddr
	var errl error
	if l.SourceIP != "" {
		localAddr, errl = net.ResolveIPAddr("ip", l.SourceIP)
	} else {
		localAddr, errl = net.ResolveIPAddr("ip", l.IP)
	}
	if errl != nil {
		panic(errl)
	}

	localTCPAddr := net.TCPAddr{
		IP: localAddr.IP,
	}

	dialer := (&net.Dialer{
		LocalAddr: &localTCPAddr,
		Timeout:   10 * time.Second,
		KeepAlive: 10 * time.Second,
		DualStack: true,
	}).DialContext

	transport := &customTransport{
		LocalAddr: &localTCPAddr,
		Transport: &http.Transport{
			Proxy:                 proxy,
			TLSClientConfig:       tlsClientConfig,
			DialContext:           dialer,
			TLSHandshakeTimeout:   10 * time.Second,
			IdleConnTimeout:       10 * time.Second,
			MaxIdleConns:          100,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}

	// Websockets are not supported using HTTP/2, so if you use that, force HTTP/1.X
	if l.HTTPProto != 1 {
		log.Debugf("enable HTTP/2 transport")
		err := http2.ConfigureTransport(transport.Transport)
		if err != nil {
			log.Fatalf("failed to prepare transport for HTTP/2: %v", err)
		}
	}

	reverseproxy := &httputil.ReverseProxy{
		Director:       director,
		Transport:      transport,
		ModifyResponse: modifyresponse,
		FlushInterval:  250 * time.Millisecond, // good for streams adn server-sent events
		ErrorLog:       logging.StandardLog("httpproxy/reverseproxy"),

		// ErrorLog:        logrus not supported as logging until v2 => https://github.com/golang/go/issues/21082
	}
	return reverseproxy
}
