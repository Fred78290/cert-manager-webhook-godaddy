#!/bin/bash
sudo rm -rf out

VERSION=v1.26.1
REGISTRY=devregistry.aldunelabs.com

make -e REGISTRY=$REGISTRY -e TAG=$VERSION container-push-manifest
