#!/usr/bin/env bash
PROJECT=trendyol-consul-template-injector

: ${1?'missing key directory'}

key_dir="$1"

chmod 0700 "$key_dir"
cd "$key_dir"

# Generate the CA cert and private key
openssl req -nodes -new -x509 -keyout ca.key -out ca.crt -subj "/CN=trendyol-consul-template-injector"

# Generate the private key for consul-template-injector
openssl genrsa -out $PROJECT-tls.key 2048

# Generate a Certificate Signin Request(CSR) for the private key
openssl req -new -key $PROJECT-tls.key -subj "/CN=trendyol-consul-template-injector-server-service.admission.svc" -out $PROJECT.csr

# Sign it with private key of the CSA
openssl x509 -req -in $PROJECT.csr -CA ca.crt -CAkey ca.key -CAcreateserial \
-out $PROJECT-tls.crt

