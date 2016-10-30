# Monitor

To run it from docker

```bash
$ docker run -t -e HOSTNAME=laptop -e INTERVAL=1 -e LEVEL=INFO -e LOGSTASH_HTTP_ENDPOINT=http://192.168.99.100:31311 -e "VOLUMES=/ /tmp" hbouvier/go-monitor
```