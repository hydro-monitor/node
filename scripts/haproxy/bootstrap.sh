#! /bin/bash

set -exuo pipefail

# Install HA Proxy
apt install haproxy

cat > /etc/haproxy/haproxy.cfg <<EOF
global
	debug
	log /dev/log local0
	log /dev/log local1 notice
	chroot /var/lib/haproxy
	stats socket /run/haproxy/admin.sock mode 660 level admin
	stats timeout 30s
	user haproxy
	group haproxy
	daemon

	# Default SSL material locations
	ca-base /etc/ssl/certs
	crt-base /etc/ssl/private

	# Default ciphers to use on SSL-enabled listening sockets.
	# For more information, see ciphers(1SSL). This list is from:
	#  https://hynek.me/articles/hardening-your-web-servers-ssl-ciphers/
	# An alternative list with additional directives can be obtained from
	#  https://mozilla.github.io/server-side-tls/ssl-config-generator/?server=haproxy
	ssl-default-bind-ciphers ECDH+AESGCM:DH+AESGCM:ECDH+AES256:DH+AES256:ECDH+AES128:DH+AES:RSA+AESGCM:RSA+AES:!aNULL:!MD5:!DSS
	ssl-default-bind-options no-sslv3

defaults
	log	global
	mode	http
	option	httplog
	option	dontlognull
        timeout connect 5000
        timeout client  50000
        timeout server  50000
	errorfile 400 /etc/haproxy/errors/400.http
	errorfile 403 /etc/haproxy/errors/403.http
	errorfile 408 /etc/haproxy/errors/408.http
	errorfile 500 /etc/haproxy/errors/500.http
	errorfile 502 /etc/haproxy/errors/502.http
	errorfile 503 /etc/haproxy/errors/503.http
	errorfile 504 /etc/haproxy/errors/504.http
	retries 3

frontend main
	bind *:80
	default_backend app

backend app
	balance roundrobin
	option httpchk GET /health
	option log-health-checks
	server srv1 104.211.23.145:8080 check port 8080 inter 120000 fall 3
	server srv2 13.90.231.210:8080 check port 8080 inter 120000 fall 3
	server srv3 18.229.102.224:443 check port 443 inter 120000 fall 3
EOF

# Restart HA Proxy with new config
systemctl restart haproxy

# HAP logs will be in /var/log/haproxy.log file