ALL_ARCH = amd64 arm64

.EXPORT_ALL_VARIABLES:

all: $(addprefix build-arch-,$(ALL_ARCH))

VERSION_MAJOR ?= 1
VERSION_MINOR ?= 26
VERSION_BUILD ?= 0
TAG?=$(VERSION_MAJOR).$(VERSION_MINOR).$(VERSION_BUILD)
KUBE_VERSION=1.26.0
FLAGS=
ENVVAR=
GOOS?=$(shell go env GOOS)
GOARCH?=$(shell go env GOARCH)
REGISTRY?=fred78290
BUILD_DATE?=`date +%Y-%m-%dT%H:%M:%SZ`
VERSION_LDFLAGS=-X main.phVersion=$(TAG)
IMAGE=$(REGISTRY)/cert-manager-godaddy

$(shell mkdir -p "$(OUT)")

export TEST_ASSET_ETCD=_test/kubebuilder/bin/etcd
export TEST_ASSET_KUBE_APISERVER=_test/kubebuilder/bin/kube-apiserver
export TEST_ASSET_KUBECTL=_test/kubebuilder/bin/kubectl
export TEST_MANIFEST_PATH=_test/kubebuilder/godaddy

test: _test/kubebuilder
	go test -v .

_test/kubebuilder:
	./scripts/config.sh https://go.kubebuilder.io/test-tools/$(KUBE_VERSION)/$(GOOS)/$(GOARCH)

deps:
	go mod vendor

build: $(addprefix build-arch-,$(ALL_ARCH))

build-arch-%: deps clean-arch-%
	$(ENVVAR) GOOS=$(GOOS) GOARCH=$* go build -ldflags="-X main.phVersion=$(TAG) -X main.phBuildDate=$(BUILD_DATE)" -a -o out/$(GOOS)/$*/cert-manager-godaddy

make-image: $(addprefix make-image-arch-,$(ALL_ARCH))

make-image-arch-%:
	docker build --pull -t ${IMAGE}-$*:${TAG} -f Dockerfile.$* .
	@echo "Image ${TAG}-$* completed"

push-image: $(addprefix push-image-arch-,$(ALL_ARCH))

push-image-arch-%:
	docker push ${IMAGE}-$*:${TAG}

push-manifest:
	docker buildx build --pull --platform linux/amd64,linux/arm64 --push -t ${IMAGE}:${TAG} .
	@echo "Image ${TAG}* completed"

container-push-manifest: container push-manifest

clean: $(addprefix clean-arch-,$(ALL_ARCH))

clean-arch-%:
	rm -f ./out/$(GOOS)/$*/cert-manager-godaddy

docker-builder:
	test -z "$(docker image ls | grep cert-manager-godaddy-builder)" && docker build -t cert-manager-godaddy-builder ./builder

build-in-docker: $(addprefix build-in-docker-arch-,$(ALL_ARCH))

build-in-docker-arch-%: clean-arch-% docker-builder
	docker run --rm -v `pwd`:/gopath/src/github.com/Fred78290/cert-manager-webhook-godaddy/ cert-manager-godaddy-builder:latest bash \
		-c 'cd /gopath/src/github.com/Fred78290/cert-manager-webhook-godaddy  \
		&& BUILD_TAGS=${BUILD_TAGS} make -e REGISTRY=${REGISTRY} -e TAG=${TAG} -e BUILD_DATE=`date +%Y-%m-%dT%H:%M:%SZ` build-arch-$*'

container: $(addprefix container-arch-,$(ALL_ARCH))

container-arch-%: build-in-docker-arch-%
	@echo "Full in-docker image ${TAG}-$* completed"

go-lint:
	golangci-lint run --timeout=15m ./...

format:
	test -z "$$(find . -path ./vendor -prune -type f -o -name '*.go' -exec gofmt -s -d {} + | tee /dev/stderr)" || \
    test -z "$$(find . -path ./vendor -prune -type f -o -name '*.go' -exec gofmt -s -w {} + | tee /dev/stderr)"


.PHONY: all deps build clean format execute-release dev-release docker-builder build-in-docker release generate

