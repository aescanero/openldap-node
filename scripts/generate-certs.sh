#!/bin/bash
#
# Script to generate self-signed certificates for HTTPS testing
# Copyright [2024] Alejandro Escanero Blanco
#

set -e

# Create certs directory if it doesn't exist
CERTS_DIR="${1:-./certs}"
mkdir -p "$CERTS_DIR"

echo "Generating self-signed certificates in $CERTS_DIR..."

# Generate private key
openssl genrsa -out "$CERTS_DIR/server.key" 2048

# Generate certificate signing request
openssl req -new -key "$CERTS_DIR/server.key" \
    -out "$CERTS_DIR/server.csr" \
    -subj "/C=ES/ST=Madrid/L=Madrid/O=DisasterProject/OU=OpenLDAP/CN=localhost"

# Generate self-signed certificate (valid for 365 days)
openssl x509 -req -days 365 \
    -in "$CERTS_DIR/server.csr" \
    -signkey "$CERTS_DIR/server.key" \
    -out "$CERTS_DIR/server.crt" \
    -extfile <(printf "subjectAltName=DNS:localhost,DNS:127.0.0.1,IP:127.0.0.1")

# Set permissions
chmod 600 "$CERTS_DIR/server.key"
chmod 644 "$CERTS_DIR/server.crt"

echo "✓ Certificates generated successfully!"
echo ""
echo "Files created:"
echo "  - Certificate: $CERTS_DIR/server.crt"
echo "  - Private Key: $CERTS_DIR/server.key"
echo "  - CSR (can be deleted): $CERTS_DIR/server.csr"
echo ""
echo "To use these certificates, update your config.yaml with:"
echo "  api_tls:"
echo "    enabled: true"
echo "    cert_file: $CERTS_DIR/server.crt"
echo "    key_file: $CERTS_DIR/server.key"
echo "    port: \"9443\""
echo "    auto_redirect: true"
echo ""
echo "⚠️  Note: These are self-signed certificates for testing only!"
echo "   For production, use certificates from a trusted CA."
