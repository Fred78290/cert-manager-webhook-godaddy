<p align="center">
  <img src="./images/cert-manager-godaddy.svg" height="256" width="256" alt="cert-manager-webhook-godaddy project logo" />
</p>

<p align="center">
<a href="https://github.com/Fred78290/cert-manager-webhook-godaddy/actions/workflows/ci.yaml">
  <img alt="Build Status" src="https://github.com/Fred78290/cert-manager-webhook-godaddy/actions/workflows/ci.yaml/badge.svg?branch=master">
</a>
<a href="https://sonarcloud.io/dashboard?id=Fred78290_cert-manager-webhook-godaddy">
  <img alt="Quality Gate Status" src="https://sonarcloud.io/api/project_badges/measure?project=Fred78290_cert-manager-webhook-godaddy&metric=alert_status">
</a>
<a href="https://github.com/Fred78290/cert-manager-webhook-godaddy/blob/master/LICENSE">
  <img alt="Licence" src="https://img.shields.io/hexpm/l/plug.svg">
</a>
</p>

# Time to leave GoDaddy...

**QUESTION: GoDaddy ACCESS DENIED via API-Call**
```
Hi,

We have recently updated the account requirements to access parts of our production Domains API. As part of this update, access to these APIs are now limited:

    Availability API: Limited to accounts with 50 or more domains
    Management and DNS APIs: Limited to accounts with 10 or more domains and/or an active Discount Domain Club plan.

If you have lost access to these APIs, but feel you meet these requirements, please reply back with your account number and we will review your account and whitelist you if we have denied you access in error.

Please note that this does not affect your access to any of our OTE APIs.

If you have any further questions or need assistance with other API questions, please reach out.

Regards,

API Support Team
```


# cert-manager webhook for GoDaddy

## Installation

```bash
helm repo add godaddy-webhook https://fred78290.github.io/cert-manager-webhook-godaddy/
helm repo update

helm upgrade -i godaddy-webhook godaddy-webhook/godaddy-webhook \
    --set groupName=acme.mycompany.com \
    --set image.tag=v1.27.2 \
    --set image.pullPolicy=Always \
    --namespace cert-manager
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
        - '*.mycompany.com'
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
