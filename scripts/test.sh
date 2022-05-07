#!/bin/bash
CURDIR=$(dirname $0)

GOOS=$(go env GOOS)
GOARCH=$(go env GOARCH)
GODADDY_API_KEY_BASE64=$(echo -n "$GODADDY_API_KEY" | base64)
GODADDY_API_SECRET_BASE64=$(echo -n "$GODADDY_API_SECRET" | base64)
KUBE_VERSION=1.23.5

pushd $CURDIR/../

export TEST_ASSET_ETCD=_test/kubebuilder/bin/etcd
export TEST_ASSET_KUBE_APISERVER=_test/kubebuilder/bin/kube-apiserver
export TEST_ASSET_KUBECTL=_test/kubebuilder/bin/kubectl
export TEST_MANIFEST_PATH=_test/kubebuilder/godaddy
export TEST_ZONE_NAME=aldunelabs.com

mkdir -p $TEST_MANIFEST_PATH

curl -fsSL https://go.kubebuilder.io/test-tools/${KUBE_VERSION}/${GOOS}/${GOARCH} -o kubebuilder-tools.tar.gz

mkdir -p _test/kubebuilder

pushd _test
tar -xvf ../kubebuilder-tools.tar.gz
popd

rm kubebuilder-tools.tar.gz

cat > $TEST_MANIFEST_PATH/api-key.yaml <<EOF
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

cat > $TEST_MANIFEST_PATH/config.json <<EOF
{
  "apiKeySecretRef": {
    "key": "$GODADDY_API_KEY",
    "secret": "$GODADDY_API_SECRET"
  },
  "production": true,
  "ttl": 600
}
EOF

exit

TEST_ZONE_NAME="${TEST_ZONE_NAME}." TEST_MANIFEST_PATH=$TEST_MANIFEST_PATH go test .

popd