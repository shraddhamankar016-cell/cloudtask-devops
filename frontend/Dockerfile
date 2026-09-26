FROM nginx:1.27-alpine

# Static assets
COPY index.html style.css script.js config.template.js /usr/share/nginx/html/

# Nginx config template (rendered at container start-up)
RUN mkdir -p /etc/nginx/templates
COPY nginx.conf.template /etc/nginx/templates/nginx.conf.template

# Entrypoint that renders templates using env vars, then starts nginx
COPY docker-entrypoint.sh /docker-entrypoint.d/40-render-config.sh
RUN chmod +x /docker-entrypoint.d/40-render-config.sh

ENV API_URL="/api"
ENV BACKEND_HOST="cloudtask-backend"

EXPOSE 80

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD wget -qO- http://localhost/health || exit 1

ENTRYPOINT ["/docker-entrypoint.d/40-render-config.sh"]
