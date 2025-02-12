# 1.0.2

misc: fix: better cname handling with latest resolver update (#116)

* update deps to get improved resolver, aswell as improvements in other dependencies

* fix websocket test with new parameters

updating change log with latest commit

# 1.0.1

misc: update readme for docker containers

group docker files in dir, and fix version reporting of docker containers (#115)


set git details even if we don't update change log

# 1.0.0

misc: After 2 years of running in production, its time to bump Mercury to a stable version number: 1.0.0 
      this version is fully backwards compatible with the previous version, and is only a cosmetic version change.

# 0.15.8

feat: TLS 1.3 is now enabled by default on all tls listeners
misc: add security policy (#113)
misc: add docker files, examples and update documentation (#111)
fix: changes to fix tests in golang 1.13 - params not in init anymore

# 0.15.7

misc: fix: prevent paging in ci scripts (#110)
misc: fix codacy issue

# 0.15.5

misc: fix: bump cookbook on new release (#106)
misc: improve readme

# 0.15.4

misc: fix: correct copied dots in ci (#104)

# 0.15.1
Important: X-Forwarded-for now gets automaticly added, if you have X-forwarded-for headers in place, you can remove them. if not, you will have multiple entries in your X-Forwarded-for header

Feature:
  * Show golang version used to compile mercury

Bug:
  * Security fixes in the golang net/http service (recompiled with latest golang vesion)

# 0.15.0
Feature:
  * Allow passing of client certificate using ACLs (lucaslorentz)
  * Adding a sourceip at the listener pool, will force all outgoing connections on that pool to use that as a source ip (use this to prevent martians when hosts have multiple network routes)
  * Added experimental support for TLS1.3 (enabled using GODEBUG=tls13=1, this will be enabled by default when go1.13 is released)

Bug:
  * Default netmask for ipv6 is now /128 instead of /32 (affects network interfaces and topology based loadbalancing checks)

Change:
  * Removed manual websocket implementation for httputil's reverse proxy. it now supports websockets in the main branch
  * DNS forwarder library replaced with one that correctly resolves cnames recursively
  * Error logging of the http/reverseproxy server is now logged in the log file instead of stderr


# 0.14.4
Feature:
  * Improved logging on balance selection in debug mode
  * Improved detailed view of backend
  * Updated all vendor libraries linked to this project
  * Updated to go1.12

Bug:
  * Fix memory leak on non-closing channel (firstbyte received) in the TCP netPipe function used for tcp loadbalancing, when no data was received.

# 0.14.3
Bug:
  * Added a garbage collection on reload due to golang bug, also to track memory usage

# 0.14.2
Bug:
  * Fixed nil pointer issue in modify cookie

# 0.14.1
Bug:
  * Fixed the outbound acl from spawning multiple set-cookies (affected replace and addcookie)

# 0.14.0
Feature:
  * Allow checking of specific parts of the LB using --dns-name, --pool-name, --backend-name and --cluster-only to check specific backends/pools/dns or cluster parts

# 0.13.9
Bug:
  * Health check report had wrong wording when reporting 2 nodes online when 1 was expected

# 0.13.8
Feature:
  * Be able to specify the amount of cluster nodes, or backend nodes for monitoring to alert on
  * Optionally enable golang pprof profiler using the PROFILER_ADDR="localhost:6060" environment variable

# 0.13.7
Feature:
  * Support TLS Client certificates in checks

# 0.13.6
Feature:
  * Added option for REQ_QUERY in ACL url rewriting

# 0.13.5
Change:
  * Added additional security headers to web gui
  * CIDR inbound ACL's now also apply to TCP proxy

# 0.13.4
Feature:
  * Added option to rewrite url before passing it on to the backend
  * Added option to allow/deny requests based on url matches

# 0.13.3
Feature:
  * Added option to use SSH authentication as health check for a node

# 0.13.2
Change:
  * Topology based loadbalancing now returns records based on closest match (instead of all matching records)
 
# 0.13.1
Feature:
  * Add option to healthcheck to follow redirects httpget/httpposts (default is no)

Bug:
  * fix nil error if secure cookie or http only cookie was not set on a acl

# 0.13.0
Feature:
  * It is now possible to modify cookies based on specific values

Bug:
  * Fixed issue where request headers with cookie name X, would not return responses with the same cookie name
  * Fixed issue where adding cookies would not be possible if a shorter cookie name, matched the start of a longer cookie name

# 0.12.6
Bug:
  * Ensure that if a webserver sends an encoded content without a content type, that mercury does not try to detect and send a type to the client

# 0.12.5
Feature:
  * backend name is included in healthcheck page

Bug:
  * Exclicitly deny non-query opscodes, as we don't support remote updates at this point
  * Reload did not affect new ldap configuration

# 0.12.4
Bug:
  * No longer attempt to bind listener ip for dns-only loadbalancing config

# 0.12.3
Change:
  * healthcheck configuration item renames to healthchecks, as it now contains multiple healthchecks

# 0.12.2
Feature:
  * Allow filters in LDAP authentication to verify user uid after login

Bug:
  * Correct JSON unmarshal for backend check
  * Remove lingering error messages if healthcheck recovered

# 0.12.1
Feature:
  * Improve html/css validation for better browser support

Bug:
  * Fix reload issue when changing cipher or curve preferences (#44)

# 0.12.0:
Features:
  * Option to set a backend to state "Maintenance" (via healthcheck or gui)
  * Ability to set a maintenance page on a backend or pool
  * Alternative state on healthcheck (e.g. online="maintenance" or offline="online")
  * Setting a backend to "maintenance" will keep serving existing connections, but no longer accept new connections

Bug:
  * Fix dns name/topology changes not taking effect on reload of config
  * Fix CSRF issue in GUI
  * Fix dns return code on non-existing hostname

# 0.11.1:
Change:
  * Improved logging messages

Bug:
  * Fix CSS for node/ip fader in backend
  * Fix default TTL at source so all interfaces show correct TTL

# 0.11.0:
Features:
  * Added HealthChecks tab for showing specific health Checks and added API calls (#12)
  * Add ability to force the health of a healthcheck using the admin GUI (#18)
  * Add LDAP and local autentication options to the GUI

Changes:
  * Adding Circle-Ci

Bug:
  * fix race condition when forming multiple clusters
  * fix double close race condition on cluster node exit
  * improve stability when 2 nodes connecting to eachother on the same milisecond

# 0.10.1:
Changes:
  * You can now specify a topology per backend Node allowing you to do topology based loadbalancing on proxy level too

Bug:
  * Default TTL on all outgoing dns requests is now set to 10 seconds

# 0.10.0:
Feature:
  * Added Support for multiple healthchecks
  * Added Support for healtchecks on VIP - these would affect all backends of the vip
  * Added Support for ICMP/UDP/TCP pings

Changes:
  * Now a random time before first health check (max 5000ms) to spread the load on servers with many checks

Bugs:
  * Reload could cause errors/state to be incorrectly displayed on multiple nodes in the same backend (gui only)
  * Fix network dependency in mercury service for systemd

# 0.9.4:
Bugs:
   * Fix issues with OCSP stapling when using SNI certificates

# 0.9.3:
Changes:
  * Cluster config has changed to increase stability within the cluster - see readme for config changes
  * Graphing to collectd has been removed, splunk is the prefered way to go. code is still in place should we change our mind

Bugs:
  * Fix incorrect listener exit on update causing crash
  * Fix certificate loading order, since map is random - causing issues on reload
  * Add correct no-caching headers to sorry and mercury custom errors
  * Fix 0x20 case insensitive requests beeing handled according to https://tools.ietf.org/html/draft-vixie-dnsext-dns0x20-00

# 0.9.2:
Bugs:
  * Fix incorrect listener exit on update causing crash
  * Fix certificate loading order, since map is random - causing issues on reload

# 0.9.2:
Feature:
  * Your now allowed to specify the amount of cluster nodes that serve a dns record

Bugs:
  * Fix loglevel not affected by reload
  * Fix monitoring message output to correctly show only the failing nodes on glb errors

# 0.9.1:
Feature:
  * Per Backend Error/Sorry page can now be specified

Bugs:
  * No longer send SRVFAIL on non-existing AAAA records, if a A record does exists.
  * Fixed possible index out of range issue in healtcheck on reload
  * Fixed crash when requesting a dns without a domain

# 0.9.0:
Changes:
  * IMPORTANT! Removed cross-connects - instead add multiple nodes to both backends for stickyness that supports proper failover
  * UUID's are now hash based, so they won't change up on restarts

Bugs:
  * Fix locking issue that could occur on dns updates
  * Fix possible dns pointer overwrite before cluster updates were sent

# 0.8.9:
Feature:
 * Add option to specify at which level to trigger sorry page, will always trigger on internal errors, but you can specify to trigger on 500+ or other result codes
 * Add OCSP Stapling support for SSL certificate verification (enabled by default for all https sites)
 * Add option to deny requests based on header match
 * Add option to allow/deny request based on CIDR

# 0.8.8:
Feature:
 * Added firstavailable loadbalancing type. this returns only 1 host if multiple are availble. usable for compatibility reasons if needed.
 * Added option to use vip in active/passive setup - this is used by monitoring only: will alert if 0 or >1 nodes/pools are online

Bugs:
 * Correct alerting on offline GLB entries

# 0.8.7:
Changes:
 * Only setup a proxy if there is a listener IP, otherwise treat it as a dns balancer only

# 0.8.6:
Feature:
 * Now supports DNS forwarding for specified cidr's

Changes:
 * Allow resolving and serving of domain-only A and CNAME records

Bugs:
 * fix dead channels for configurations where the proxy function is disabled
 * fix dns authoritive and recursive answers in replies
 * fix locking issue on race condition during startup
 * fix crash that could occur on invalid dns request
 * correct dns reply return codes
 * fix content length issue on error page

# 0.8.5:

Changes:
 * Offline GLB pools now return all IP's instead of none, directing client to proper error instead of dns not found

Bugs:
 * fix incorrect domain name in dns result
 * fix backend duplication on reload with cross-connects with more than one node in a single backend

# 0.8.4:
Features:
 * Websocket support (note that you must force httpproto on the listener to 1, as websocket is not supported by http/2 which is enabled by default)
 * Better support for SOA records and serial updating

Bugs:
 * Stale node in proxy config if removed by reload, should no longer occur
 * Properly handle main and sub certificates and check them all during config loading
 * Fix web interface for local dns entries
 * Fix additional replies for dns entries

# 0.8.3:
Features:
 * Origin traffic is now the listener ip for both proxy and healthcheck
 * Better deal with timeouts in healthcheck

Bugs:
 * Fix healtcheck json unmarshal for duration for backend check
 * Fix healtcheck json unsupported type: chan bool
 * Fix healtcheck don't report internal vip's as down, they are always up
 * Fix sticky session if pointer no longer exists
 * Remove deadline timeout on tcp proxy

# 0.8.2:
Features:
 * DNS server now uses proxy statistics for loadbalancing algorithm when using internal loadbalancer (uses its own counter if not)

Bugs:
 * Fix Roundrobin statistics

# 0.8.1:
Features:
 * Change health check parameters to be more clear on check type
    * config file changes:
    * reply -> httpreply / tcpreply depending on check
    * request -> httprequest / tcprequest depending on check
    * postdata -> httppostdata
 * No longer are session ID's automaticly added
    * Requires config to add:
    * example: { action: 'add', cookie_key: 'mercid', cookie_value: '##UUID##', cookie_expire: '24h', cookie_secure: true, cookie_httponly: true }
 * Sticky session cookies are only parsed if we use sticky based loadbalancing
 * adding of cookies will now only add if cookie is not set yet

Bugs:
 * fix ACLs on self-generated responses

# 0.8.0:
Features:
 * HTTP/2 Support added for both client and backend
 * ResponseTime based Load-balancing added
 * Failed http requests to backend now return 500 Internal Server Error

Bugs:
 * DNS responses now are case insensitive
 * Fix client connected count

# 0.7.0:
Features:
 * Add sorry page abilit
 * Added client session tracking

Bugs:
 * Reload now works for DNS Listene
 * Fix concurrency issue

# 0.6.0:
 * Start of Change log
