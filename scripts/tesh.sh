#!/bin/bash
CURDIR=$(dirname $0)

GODADDY_API_KEY_BASE64=$(echo -n "$GODADDY_API_KEY" | base64)
GODADDY_API_SECRET_BASE64=$(echo -n "$GODADDY_API_SECRET" | base64)

pushd $CURDIR/../

mkdir -p __main__/testdata/godaddy

cat > __main__/testdata/godaddy/api-key.yaml <<EOF
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

cat > __main__/testdata/godaddy/config.json <<EOF
{
  "apiKeySecretRef": {
    "key": "$GODADDY_API_KEY",
    "secret": "$GODADDY_API_SECRET"
  },
  "production": true,
  "ttl": 600
}
EOF

TEST_ZONE_NAME=aldunelabs.com. TEST_MANIFEST_PATH=__main__/testdata/godaddy go test .

popd