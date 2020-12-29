#!/bin/bash

make -e REGISTRY=fred78290 -e TAG=v1.0.0 container

docker tag fred78290/cert-manager-godaddy:v1.0.0 localhost:32000/cert-manager-godaddy:v1.0.0
docker push localhost:32000/cert-manager-godaddy:v1.0.0
