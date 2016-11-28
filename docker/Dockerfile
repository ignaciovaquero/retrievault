FROM phusion/baseimage:0.9.19

# TODO: Set labels
RUN mkdir -p /etc/my_init.d && \
    mkdir -p /etc/retrievault/config && \
    mkdir -p /etc/retrievault/generic && \
    mkdir -p /etc/retrievault/ca-cert && \
    mkdir -p /etc/retrievault/certs

RUN echo "/var/log/retrievault.log" > /etc/container_environment/LOG_FILE && \
    echo "/etc/retrievault/config/config.json" > /etc/container_environment/CONFIG_FILE

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
