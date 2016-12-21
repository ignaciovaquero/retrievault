FROM alpine:3.4

LABEL org.label-schema.schema-version="1.0" \
      org.label-schema.name="retrievault" \
      org.label-schema.description="Retrievault is a container which retrieves secrets from vault and saves them to files" \
      org.label-schema.vcs-url="https://github.com/DatioBD/retrievault" \
      org.label-schema.vcs-ref="c1a251d" \
      org.label-schema.vendor="Datio Big Data" \
      org.label-schema.version="0.2.3-SNAPSHOT" \
      org.label-schema.usage="https://github.com/DatioBD/retrievault/blob/master/README.md" \
      org.label-schema.docker.cmd="docker run -d -e VAULT_TOKEN=<your_token> -v /path/to/vault/ca.crt:/usr/local/share/ca-certificates/ca.crt -v $(pwd)/config.json:/etc/retrievault/config/config.json eu.gcr.io/datio-sistemas/retrievault:0.1.0-SNAPSHOT"

ARG DUMB_INIT_VERSION=1.2.0

RUN apk update && apk add vim=7.4.1831-r2 \
                          ca-certificates=20160104-r4 \
                          curl=7.51.0-r0 \
                          openssl=1.0.2j-r0 && \
    curl -sSL https://github.com/Yelp/dumb-init/releases/download/v${DUMB_INIT_VERSION}/dumb-init_${DUMB_INIT_VERSION}_amd64 > /usr/local/bin/dumb-init && \
    chmod 0755 /usr/local/bin/dumb-init && \
    mkdir -p /etc/retrievault/config && \
    mkdir -p /etc/retrievault/generic && \
    mkdir -p /etc/retrievault/ca-cert && \
    mkdir -p /etc/retrievault/certs

ADD retrievault /usr/bin/retrievault
ADD entrypoint.sh /entrypoint.sh

VOLUME /usr/local/share/ca-certificates \
       /etc/retrievault/config \
       /etc/retrievault/certs \
       /etc/retrievault/ca-cert \
       /etc/retrievault/generic

ENTRYPOINT ["/usr/bin/dumb-init", "--"]
CMD ["/entrypoint.sh"]
