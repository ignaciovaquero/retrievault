#!/bin/sh

cp -a /usr/local/share/ca-certificates/*.crt /etc/ssl/certs/
/usr/sbin/update-ca-certificates && \
/usr/bin/retrievault
