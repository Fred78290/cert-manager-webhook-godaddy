[![Build status](https://github.com/Fred78290/cert-manager-webhook-godaddy/actions/workflows/ci.yaml/badge.svg?branch=master)](https://github.com/Fred78290/cert-manager-webhook-godaddy/actions/workflows/ci.yaml)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=Fred78290/cert-manager-webhook-godaddy&metric=alert_status)](https://sonarcloud.io/dashboard?id=Fred78290_cert-manager-webhook-godaddy)
[![Licence](https://img.shields.io/hexpm/l/plug.svg)](https://github.com/Fred78290/cert-manager-webhook-godaddy/blob/master/LICENSE)
# ACME webhook for GoDaddy


## Installation

```bash
helm install godaddy-webhook \
    --set groupName=acme.mycompany.com \
    --set image.repository=fred78290/cert-manager-godaddy \
    --set image.tag=v1.20.5 \
    --set image.pullPolicy=Always \
    --namespace cert-manager ./deploy/godaddy-webhook
```

## Issuer

### ClusterIssuer

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: godaddy-api-key-prod
  namespace: cert-manager
type: Opaque
data:
  key: <godaddy api key base64 encoded>
  secret: <godaddy api secret base64 encoded>
---  
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: <your email>
    privateKeySecretRef:
      name: letsencrypt-prod-account-key
    solvers:
    - selector:
        dnsNames:
        - '*.example.com'
      dns01:
        webhook:
          config:
            apiKeySecretRef:
              name: godaddy-api-key-prod
              key: key
              secret: secret
            production: true
            ttl: 600
          groupName: acme.mycompany.com
          solverName: godaddy
```

Certificate

```yaml
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: wildcard-example-com
spec:
  secretName: wildcard-example-com-tls
  renewBefore: 240h
  dnsNames:
  - '*.example.com'
  issuerRef:
    name: letsencrypt-prod
    kind: ClusterIssuer
```

Ingress

```yaml
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: example-ingress
  namespace: default
  annotations:
    certmanager.k8s.io/cluster-issuer: "letsencrypt-prod"
spec:
  tls:
  - hosts:
    - '*.example.com'
    secretName: wildcard-example-com-tls
  rules:
  - host: demo.example.com
    http:
      paths:
      - path: /
        backend:
          serviceName: backend-service
          servicePort: 80
```

## Development

### Running the test suite
All DNS providers **must** run the DNS01 provider conformance testing suite,
else they will have undetermined behaviour when used with cert-manager.

**It is essential that you configure and run the test suite when creating a
DNS01 webhook.**

An example Go test file has been provided in [main_test.go]().

> Prepare

```bash
$ scripts/fetch-test-binaries.sh
```

You can run the test suite with:

```bash
$ scripts/test.sh
```

The example file has a number of areas you must fill in and replace with your
own options in order for tests to pass.
