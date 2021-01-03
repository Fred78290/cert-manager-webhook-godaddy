#!/bin/bash

make -e REGISTRY=fred78290 -e TAG=v1.18.14 container

docker tag fred78290/cert-manager-godaddy:v1.18.14 localhost:32000/cert-manager-godaddy:v1.18.14
docker push localhost:32000/cert-manager-godaddy:v1.18.14
