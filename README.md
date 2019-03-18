# ACME.SH.WS

Webservice wrapper for acme.sh

## Installation

### Webservice

```
 docker run --rm \
 -v $(pwd)/home:/acme.sh \
 -v $(pwd)/certs:/certs \
 -v $(pwd)/webroot:/webroot \
 -e DEPLOY_HOOK=haproxy \
 -e DEPLOY_HAPROXY_PEM_PATH=/certs \
 -e CF_Key="XXXXX" \
 -e CF_Email="XXXX" \
 -e WEBROOT_DIR=/webroot
 vixns/acme.sh.ws
```

### Auto Renew process

Use a scheduler (cron/chronos/...) to run this command once a day

```
 docker run --rm \
 -v $(pwd)/home:/acme.sh \
 -v $(pwd)/certs:/certs \
 -e DEPLOY_HOOK=haproxy \
 -e DEPLOY_HAPROXY_PEM_PATH=/certs \
 -e CF_Key="XXXXX" \
 -e CF_Email="XXXX" \
 -e WEBROOT_DIR=/webroot
 --entrypoint=acme.sh \
 vixns/acme.sh.ws \
 --cron --home /acme.sh --auto-upgrade 0
```

## Environment

name | default | description
BIND_IP | 0.0.0.0 | IP address to bind, you should let the default when using the docker image
BIND_PORT | 3000 | port to bind
ACME_SH_PATH | /usr/local/bin/acme.sh | acme.sh binary  path
WEBROOT_DIR | | webroot dir, no default, optional if you only use dns api
DEPLOY_HOOK | | value of acme.sh --deploy-hook argument, required
DRY_RUN | | run in test mode if not empty

## Usage

### Issue a certificate using webroot mode

Your local webserver must be configured to serve `$WEBROOT_DIR/.well-known/challenges` as `http://_all_domains_and_aliases_/.well-known/challenges`.

`curl -XPOST http://localhost:3000 -d '{"domains":["my.domain.com","alias.domain.com","alias2.domain.com"]}'`

### Issue a certificate using the dns api

Documentation: https://github.com/Neilpang/acme.sh/blob/master/dnsapi/README.md

`curl -XPOST http://localhost:3000 -d '{"domains":["my.domain.com"],"dns_api":"cf"}'`

### Issue a certificate using a dns alias mode

Documentation: https://github.com/Neilpang/acme.sh/wiki/DNS-alias-mode

`curl -XPOST http://localhost:3000 -d '{"domains":["*.domain.com"],"dns_api":"cf", "challenge_alias": "myacmedomain.com"}'`

### Delete a certificate

`curl -XDELETE http://localhost:3000 -d '{"domains":["my.domain.com"]}'`

