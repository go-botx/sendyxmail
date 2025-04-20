#!/bin/sh

EXTRA_CERTS_DIR=${EXTRA_CERTS_DIR:-/certs}

# Add custom CAs if specified
if [ -n "$(ls -A "${EXTRA_CERTS_DIR}")" ]; then
    cp "${EXTRA_CERTS_DIR}/*" /etc/ssl/certs/
    update-ca-certificates
fi

exec "$@"