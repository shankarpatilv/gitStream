FROM grafana/grafana-oss:11.1.4

USER root

COPY deploy/fly/grafana/provisioning /etc/grafana/provisioning
COPY grafana/dashboard.json /tmp/gitstream-dashboard.json

RUN sed 's/${DS_PROMETHEUS}/gitstream-prometheus/g' \
  /tmp/gitstream-dashboard.json \
  > /etc/grafana/provisioning/dashboards/gitstream-pipeline.json && \
  chown -R grafana:root /etc/grafana/provisioning

USER grafana
