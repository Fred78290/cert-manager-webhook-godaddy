#!/bin/bash

helm install godaddy-webhook \
    --set groupName=aldunelabs.com \
    --set image.repository=devregistry.aldunelabs.com/cert-manager-godaddy \
    --set image.tag=v1.20.5 \
    --set image.pullPolicy=Always \
    --namespace cert-manager ./deploy/godaddy-webhook $@