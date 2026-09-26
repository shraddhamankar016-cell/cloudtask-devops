#!/bin/sh
# Renders config.template.js and nginx.conf.template using environment
# variables, then hands off to Nginx. This lets one built image be
# pointed at any backend endpoint (local, staging, prod) without a
# rebuild — a common pattern for "build once, deploy many times".
set -e

: "${API_URL:=/api}"
: "${BACKEND_HOST:=cloudtask-backend}"
export API_URL BACKEND_HOST

envsubst '${API_URL}' < /usr/share/nginx/html/config.template.js > /usr/share/nginx/html/config.js
envsubst '${BACKEND_HOST}' < /etc/nginx/templates/nginx.conf.template > /etc/nginx/conf.d/default.conf

exec nginx -g 'daemon off;'
