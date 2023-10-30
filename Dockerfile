FROM alpine:3.18
RUN apk upgrade --no-cache && \
    apk --no-cache add ca-certificates git openssh-client tini
COPY mefi /usr/bin/
USER nobody
ENTRYPOINT ["/sbin/tini", "--", "/usr/bin/mefi"]