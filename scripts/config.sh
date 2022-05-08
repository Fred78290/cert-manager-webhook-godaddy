#!/bin/bash
CURDIR=$(dirname $0)

GODADDY_API_KEY_BASE64=$(echo -n "$GODADDY_API_KEY" | base64)
GODADDY_API_SECRET_BASE64=$(echo -n "$GODADDY_API_SECRET" | base64)

curl -fsSL "$1" -o kubebuilder-tools.tar.gz

mkdir -p _test/kubebuilder/godaddy
cd _test
tar -xvf ../kubebuilder-tools.tar.gz
cd ..
rm kubebuilder-tools.tar.gz

cat > _test/kubebuilder/godaddy/api-key.yaml <<EOF
apiVersion: v1
kind: Secret
metadata:
  name: godaddy-api-key
  namespace: basic-present-record
type: Opaque
data:
  key: "$GODADDY_API_KEY_BASE64"
  secret: "$GODADDY_API_SECRET_BASE64"
EOF

cat > _test/kubebuilder/godaddy/config.json <<EOF
{
  "apiKeySecretRef": {
    "key": "$GODADDY_API_KEY",
    "secret": "$GODADDY_API_SECRET"
  },
  "production": true,
  "ttl": 600
}
EOF
