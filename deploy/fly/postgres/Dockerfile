FROM postgres:15

COPY deploy/fly/postgres/start.sh /usr/local/bin/gitstream-postgres-start
RUN chmod +x /usr/local/bin/gitstream-postgres-start

ENTRYPOINT ["gitstream-postgres-start"]
