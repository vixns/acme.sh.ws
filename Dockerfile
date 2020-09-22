FROM golang:alpine
COPY main.go .
RUN apk add --no-cache git && go get github.com/gorilla/mux && go build -o acme.sh.ws

FROM neilpang/acme.sh

ENV HOME=/root WEBROOT_DIR=/webroot DEPLOY_HOOK=haproxy FABIO=1 VAULT_VERSION=1.5.3

RUN apk add --no-cache bind-tools bash unzip && \
curl -sLO https://releases.hashicorp.com/vault/${VAULT_VERSION}/vault_${VAULT_VERSION}_linux_amd64.zip && \
unzip -q -d /usr/local/bin vault_${VAULT_VERSION}_linux_amd64.zip && rm vault_${VAULT_VERSION}_linux_amd64.zip


COPY vault_cli.sh /root/.acme.sh/deploy/vault_cli.sh
COPY --from=0 /go/acme.sh.ws /bin/acme.sh.ws
ENTRYPOINT ["/bin/acme.sh.ws"]
