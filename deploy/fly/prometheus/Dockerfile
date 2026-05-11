FROM prom/prometheus:v2.54.1

COPY deploy/fly/prometheus/prometheus.yml /etc/prometheus/prometheus.yml.template
COPY deploy/fly/prometheus/start.sh /usr/local/bin/gitstream-prometheus-start

USER root
RUN chmod +x /usr/local/bin/gitstream-prometheus-start
USER nobody

ENTRYPOINT ["/usr/local/bin/gitstream-prometheus-start"]
