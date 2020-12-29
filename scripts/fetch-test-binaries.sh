#!/usr/bin/env bash

set -e

k8s_version=1.19.2
goarch=amd64
goos="unknown"

if [[ "$OSTYPE" == "linux-gnu" ]]; then
  goos="linux"
elif [[ "$OSTYPE" == "darwin"* ]]; then
  goos="darwin"
fi

if [[ "$goos" == "unknown" ]]; then
  echo "OS '$OSTYPE' not supported. Aborting." >&2
  exit 1
fi

mkdir -p __main__/hack
curl -sfL https://storage.googleapis.com/kubebuilder-tools/kubebuilder-tools-$k8s_version-$goos-$goarch.tar.gz | tar xvz --strip-components=1 -C __main__/hack
