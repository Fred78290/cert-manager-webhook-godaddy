# ACME webhook for GoDaddy


## Installation

```bash
$ helm install --name cert-manager-webhook-godaddy ./deploy/godaddy-webhook
```

## Issuer

secret

```bash
$ kubectl -n cert-manager create secret generic godaddy-credentials --from-literal=authAPISecret='your GoDaddy authAPISecret'
```

RBAC

```yaml
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: cert-manager-webhook-godaddy:secret-reader
rules:
  - apiGroups:
      - ''
    resources:
      - 'secrets'
    resourceNames:
      - 'godaddy-credentials'
    verbs:
      - 'get'
      - 'watch'
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: cert-manager-webhook-godaddy:secret-reader
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cert-manager-webhook-godaddy:secret-reader
subjects:
  - apiGroup: ""
    kind: ServiceAccount
    name: cert-manager-webhook-godaddy
    namespace: default
```

ClusterIssuer

```yaml
apiVersion: certmanager.k8s.io/v1alpha1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: <your email>
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
    - selector: 
        dnsNames:
        - '*.example.com'
      dns01:
        webhook:
          config:
            accessKeyId: <your GoDaddy authAPIKey>
            accessKeySecretRef:
              key: accessKeySecret
              name: godaddy-credentials
            ttl: 600
          groupName: acme.company.com
          solverName: godaddy
```

Certificate

```yaml
apiVersion: certmanager.k8s.io/v1alpha1
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
$ TEST_ZONE_NAME=example.com go test .
```

The example file has a number of areas you must fill in and replace with your
own options in order for tests to pass.
