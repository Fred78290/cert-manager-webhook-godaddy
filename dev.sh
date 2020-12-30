#!/bin/bash

helm install godaddy-webhook \
    --set groupName=aldunelabs.com \
    --set image.repository=localhost:32000/cert-manager-godaddy \
    --set image.tag=v1.19.6 \
    --set image.pullPolicy=Always \
    --namespace cert-manager ./deploy/godaddy-webhook