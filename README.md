# kube-envoy-controller
Kubernetes CRD &amp; custom controller for envoyproxy
Allows creation & configuring of envoyproxies using a kubernetes resource (kind: Envoy)

# Installing
```$ go get github.com/starizard/kube-envoy-controller
   $ go build
 ```
# Usage
```
 $ kubectl apply -f crds/ sample/
 $ ./kube-envoy-controller
 $ kubectl get envoy

```

