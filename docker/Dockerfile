FROM phusion/baseimage:0.9.19

LABEL org.label-schema.schema-version="1.0" \
      org.label-schema.name="retrievault" \
      org.label-schema.description="Retrievault is a container which retrieves secrets from vault and saves them to files" \
      org.label-schema.vcs-url="https://github.com/DatioBD/retrievault" \
      org.label-schema.vcs-ref="25e0751" \
      org.label-schema.vendor="Datio Big Data" \
      org.label-schema.version="0.2.0-SNAPSHOT" \
      org.label-schema.usage="https://github.com/DatioBD/retrievault/blob/master/README.md" \
      org.label-schema.docker.cmd="docker run -d -e VAULT_TOKEN=<your_token> -v /path/to/vault/ca.crt:/usr/local/share/ca-certificates/ca.crt -v $(pwd)/config.json:/etc/retrievault/config/config.json eu.gcr.io/datio-sistemas/retrievault:0.1.0-SNAPSHOT"

RUN mkdir -p /etc/my_init.d && \
    mkdir -p /etc/retrievault/config && \
    mkdir -p /etc/retrievault/generic && \
    mkdir -p /etc/retrievault/ca-cert && \
    mkdir -p /etc/retrievault/certs

ADD retrievault /usr/bin/retrievault
ADD update_ca_certs.sh /etc/my_init.d/0_update_ca_certs.sh
ADD retrievault.sh /etc/my_init.d/retrievault.sh
ADD output_log.sh /etc/my_init.d/zz_output_log.sh

# Clean up APT.
RUN apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

VOLUME /etc/retrievault/config \
       /etc/retrievault/certs \
       /etc/retrievault/generic

CMD ["/sbin/my_init"]
