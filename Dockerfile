FROM busybox

COPY release/bin/linux_amd64/monitor /bin

ENV HOSTNAME=container \
    INTERVAL=15 \
    LEVEL=ERROR \
    LOGSTASH_HTTP_ENDPOINT=http://logstash:31311 \
    VOLUMES=/

CMD monitor -hostname ${HOSTNAME} -interval ${INTERVAL} -level ${LEVEL} -logstash=${LOGSTASH_HTTP_ENDPOINT} ${VOLUMES}
