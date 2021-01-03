#!/bin/bash

make -e REGISTRY=fred78290 -e TAG=v1.19.6 container

docker tag fred78290/cert-manager-godaddy:v1.19.6 localhost:32000/cert-manager-godaddy:v1.19.6
docker push localhost:32000/cert-manager-godaddy:v1.19.6
