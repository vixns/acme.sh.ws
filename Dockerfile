FROM golang:alpine
COPY main.go .
RUN apk add --no-cache git && go get github.com/gorilla/mux && go build -o acme.sh.ws

FROM neilpang/acme.sh

RUN apk add --no-cache bind-tools

ENV WEBROOT_DIR=/webroot \
  DEPLOY_HAPROXY_RELOAD="curl -v http://www.monip.org" \
  DEPLOY_HAPROXY_PEM_PATH=/certs \
  DEPLOY_HOOK=haproxy

COPY --from=0 /go/acme.sh.ws /bin/acme.sh.ws
ENTRYPOINT ["/bin/acme.sh.ws"]
