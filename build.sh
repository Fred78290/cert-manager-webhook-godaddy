#!/bin/bash

make -e REGISTRY=fred78290 -e TAG=v1.20.1 container

docker tag fred78290/cert-manager-godaddy:v1.20.1 localhost:32000/cert-manager-godaddy:v1.20.1
docker push localhost:32000/cert-manager-godaddy:v1.20.1
