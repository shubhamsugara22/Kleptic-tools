# Istio Security Policies

This file contains examples for implementing security in Istio.

## Mutual TLS (mTLS)

Istio automatically enables mTLS for all service-to-service communication within the mesh.

### PeerAuthentication Example

Define mTLS mode for services.

```yaml
apiVersion: security.istio.io/v1beta1
kind: PeerAuthentication
metadata:
  name: default
  namespace: istio-system
spec:
  mtls:
    mode: STRICT  # STRICT, PERMISSIVE, or DISABLE
---
# Enable PERMISSIVE mode for specific namespace during migration
apiVersion: security.istio.io/v1beta1
kind: PeerAuthentication
metadata:
  name: default
  namespace: default
spec:
  mtls:
    mode: PERMISSIVE
```

## Authorization Policies

Control which services can communicate with each other.

### Deny All Traffic

```yaml
apiVersion: security.istio.io/v1beta1
kind: AuthorizationPolicy
metadata:
  name: deny-all
  namespace: default
spec:
  rules: []
```

### Allow Specific Service

```yaml
apiVersion: security.istio.io/v1beta1
kind: AuthorizationPolicy
metadata:
  name: allow-frontend
  namespace: default
spec:
  selector:
    matchLabels:
      app: productpage
  rules:
  - from:
    - source:
        principals:
        - "cluster.local/ns/default/sa/frontend"
    to:
    - operation:
        methods:
        - "GET"
        - "POST"
        paths:
        - "/api/v1/*"
```

### Allow Multiple Sources

```yaml
apiVersion: security.istio.io/v1beta1
kind: AuthorizationPolicy
metadata:
  name: allow-multiple-services
  namespace: default
spec:
  selector:
    matchLabels:
      app: reviews
  rules:
  - from:
    - source:
        namespaces:
        - "default"
        - "monitoring"
      source:
        principals:
        - "cluster.local/ns/default/sa/productpage"
        - "cluster.local/ns/default/sa/ratings"
    to:
    - operation:
        methods:
        - "GET"
```

### Allow with Custom Attributes

```yaml
apiVersion: security.istio.io/v1beta1
kind: AuthorizationPolicy
metadata:
  name: allow-based-on-headers
  namespace: default
spec:
  selector:
    matchLabels:
      app: reviews
  rules:
  - from:
    - source:
        principals:
        - "cluster.local/ns/default/sa/productpage"
    when:
    - key: request.headers[user]
      values:
      - "alice"
      - "bob"
    to:
    - operation:
        methods:
        - "GET"
        paths:
        - "/reviews/*"
```

## RequestAuthentication

Validate JWT tokens for workloads.

```yaml
apiVersion: security.istio.io/v1beta1
kind: RequestAuthentication
metadata:
  name: jwt-auth
  namespace: default
spec:
  jwtRules:
  - issuer: "https://example.com/"
    jwksUri: "https://example.com/.well-known/jwks.json"
    audiences:
    - "my-service"
    forwardOriginalToken: true
---
# Require valid JWT
apiVersion: security.istio.io/v1beta1
kind: AuthorizationPolicy
metadata:
  name: require-jwt
  namespace: default
spec:
  selector:
    matchLabels:
      app: api-server
  rules:
  - from:
    - source:
        requestPrincipals:
        - "https://example.com/alice"
        - "https://example.com/bob"
    to:
    - operation:
        methods:
        - "GET"
        - "POST"
```

## Certificate Management

Istio automatically manages certificates using Citadel.

### View Certificates

```bash
# Check certificate details
kubectl get secret -n default | grep prod-cert

# Decode certificate
kubectl get secret -n default prod-cert -o jsonpath='{.data.tls\.crt}' | base64 -d | openssl x509 -text -noout
```

### Monitor Certificate Expiration

```bash
# List all certificates and expiration dates
for ns in $(kubectl get ns -o jsonpath='{.items[*].metadata.name}'); do
  echo "Namespace: $ns"
  kubectl get secret -n $ns -l 'istio.io/key-and-cert' \
    -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.data.tls\.crt}{"\n"}{end}' | \
    while read name cert; do
      echo -n "$name expires: "
      echo "$cert" | base64 -d | openssl x509 -noout -enddate
    done
done
```

## Apply Security Policies

```bash
# Apply a single policy
kubectl apply -f authorization-policy.yaml

# Apply all policies
kubectl apply -f ./

# Check policies
kubectl get authorizationpolicies
kubectl get peerauthentication
kubectl get requestauthentication

# Describe a policy
kubectl describe ap allow-frontend
```

## Troubleshooting

### Check mTLS Status

```bash
# Verify mTLS is enabled
kubectl get peerauthentication -A

# Check traffic policy
kubectl describe vs <service-name>
```

### Debug Authorization Denials

```bash
# Check logs of proxy
kubectl logs <pod-name> -n default -c istio-proxy | grep "denied"

# Enable debug logging
kubectl set env deployment/<app> -n default PILOT_LOG_SCOPE=debug
```

## Best Practices

1. **Start with PERMISSIVE mode**: Enable mTLS in PERMISSIVE mode first to avoid breaking existing connections
2. **Use deny all + allow list**: Start with deny all and explicitly allow required traffic
3. **Monitor before enforcing**: Use monitoring tools to identify unexpected denials before switching to STRICT
4. **Namespace isolation**: Consider applying policies per namespace
5. **Service accounts**: Use proper service accounts for fine-grained control
