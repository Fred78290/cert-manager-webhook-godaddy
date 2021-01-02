.EXPORT_ALL_VARIABLES:

all: build
VERSION_MAJOR ?= 0
VERSION_MINOR ?= 2
VERSION_BUILD ?= 0
DEB_VERSION ?= $(VERSION_MAJOR).$(VERSION_MINOR)-$(VERSION_BUILD)
TAG?=v$(VERSION_MAJOR).$(VERSION_MINOR).$(VERSION_BUILD)
FLAGS=
ENVVAR=
GOOS?=linux
GOARCH?=amd64
REGISTRY?=fred78290
BASEIMAGE?=k8s.gcr.io/debian-base-amd64:v1.0.0
BUILD_DATE?=`date +%Y-%m-%dT%H:%M:%SZ`
VERSION_LDFLAGS=-X main.phVersion=$(TAGS)

ifdef BUILD_TAGS
  TAGS_FLAG=--tags ${BUILD_TAGS}
  PROVIDER=-${BUILD_TAGS}
  FOR_PROVIDER=" for ${BUILD_TAGS}"
else
  TAGS_FLAG=
  PROVIDER=
  FOR_PROVIDER=
endif

deps:
	go mod vendor

build:
	$(ENVVAR) GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags="-X main.phVersion=$(TAG) -X main.phBuildDate=$(BUILD_DATE)" -a -o out/cert-manager-godaddy-$(GOOS)-$(GOARCH) ${TAGS_FLAG}

make-image:
	docker build --pull --build-arg BASEIMAGE=${BASEIMAGE} \
	    -t ${REGISTRY}/cert-manager-godaddy${PROVIDER}:${TAG} .

build-binary: clean deps
	$(ENVVAR) make -e BUILD_DATE=${BUILD_DATE} -e REGISTRY=${REGISTRY} -e TAG=${TAG} -e GOOS=linux -e GOARCH=amd64 build

dev-release: build-binary execute-release
	@echo "Release ${TAG}${FOR_PROVIDER} completed"

clean:
#	sudo rm -rf out

format:
	test -z "$$(find . -path ./vendor -prune -type f -o -name '*.go' -exec gofmt -s -d {} + | tee /dev/stderr)" || \
    test -z "$$(find . -path ./vendor -prune -type f -o -name '*.go' -exec gofmt -s -w {} + | tee /dev/stderr)"

docker-builder:
	test -z "$(docker image ls | grep cert-manager-godaddy-builder)" && docker build -t cert-manager-godaddy-builder ./builder

build-in-docker: docker-builder
	docker run --rm -v `pwd`:/gopath/src/github.com/Fred78290/cert-manager-webhook-godaddy/ cert-manager-godaddy-builder:latest bash \
		-c 'cd /gopath/src/github.com/Fred78290/cert-manager-webhook-godaddy  \
		&& BUILD_TAGS=${BUILD_TAGS} make -e REGISTRY=${REGISTRY} -e TAG=${TAG} -e BUILD_DATE=`date +%Y-%m-%dT%H:%M:%SZ` build-binary'

release: build-in-docker execute-release
	@echo "Full in-docker release ${TAG}${FOR_PROVIDER} completed"

container: clean build-in-docker make-image
	@echo "Created in-docker image ${TAG}${FOR_PROVIDER}"

.PHONY: all deps build clean format execute-release dev-release docker-builder build-in-docker release generate

