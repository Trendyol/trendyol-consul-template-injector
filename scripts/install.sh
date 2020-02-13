#!/usr/bin/env bash
set -euo pipefail

PROJECT=trendyol-consul-template-injector
NAMESPACE=admission



templateDir="templates"
manifestsBasedir="manifests"
keydir="$(mktemp -d)"

# Generate keys into a temporary directory.
echo "Generating TLS keys ..."
"scripts/generate-certificates.sh" "$keydir"

# Create the TLS secret for the generated keys.
kubectl --namespace $NAMESPACE create secret tls $PROJECT-server-tls-secret \
  --cert "${keydir}/$PROJECT-tls.crt" \
  --key "${keydir}/$PROJECT-tls.key"

# Read the PEM-encoded CA certificate, base64 encode it, and replace the `${CA_PEM_B64}` placeholder in the YAML
# template with it. Then, create the Kubernetes resources.
ca_pem_b64="$(openssl base64 -A <"${keydir}/ca.crt")"
sed -e 's@${CA_PEM_B64}@'"$ca_pem_b64"'@g' <"${templateDir}/mutation-web-hook.template" |
  kubectl create -f -

kubectl apply -f $manifestsBasedir/

# Delete the key directory to prevent abuse (DO NOT USE THESE KEYS ANYWHERE ELSE).
rm -rf "$keydir"

echo "The webhook server has been deployed and configured!"
