# Kube-Envoy-Controller
Kubernetes CRD &amp; custom controller for envoyproxy

Allows creation & configuring of envoyproxies using a kubernetes resource (kind: Envoy)

# Installing
```sh
   $ go get github.com/starizard/kube-envoy-controller

   $ go build
   
```
# Usage

```sh

$ kubectl apply -f crds/ 
customresourcedefinition.apiextensions.k8s.io/envoys.example.com created
 
$ ./kube-envoy-controller
 
```
 
### In a separate shell 
 
```sh
$ kubectl apply -f sample/envoy.yaml
envoy.example.com/edge-envoy created
 
$ kubectl get envoy
NAME         AGE
edge-envoy   35s
 
$ kubectl get configmap
NAME          DATA   AGE
envoy-cfg-1   1      45s
 
$ kubectl get po
NAME                       READY   STATUS    RESTARTS   AGE
envoy-1-794d4fb667-dkwww   1/1     Running   0          57s
envoy-1-794d4fb667-jpjw8   1/1     Running   0          57s
envoy-1-794d4fb667-xc6cl   1/1     Running   0          57s
 

```


# Roadmap
- [x] Envoy CRD
- [x] Autogenerate bootstrap configmap & mount it to the envoy pods
- [x] Configure XDS 
- [ ] Automatic Sidecar Injection (Mutating Webhook)
- [ ] Implement XDS component
- [ ] Ship access log & expose prometheus metrics

