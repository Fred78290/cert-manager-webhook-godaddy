#!/bin/bash


helm install godaddy-webhook \
    --set groupName=aldunelabs.com \
    --set image.repository=devregistry.aldunelabs.com/cert-manager-godaddy \
    --set image.tag=v1.26.0 \
    --set image.pullPolicy=Always \
    --namespace cert-manager ./deploy/godaddy-webhook $@

kubectl apply -f ./release/aldunelabs.yml
kubectl apply -f ./release/certificat.yml
