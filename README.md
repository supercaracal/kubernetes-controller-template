![](https://github.com/supercaracal/kubernetes-controller-template/workflows/Test/badge.svg?branch=master)
![](https://github.com/supercaracal/kubernetes-controller-template/workflows/Release/badge.svg)

Kubernetes Controller Template
===============================================================================

This controller has a feature to log a message declared by manifest.

# Running on local host
```
$ kind create cluster
$ make apply-manifests
$ make build
$ make run
I0922 13:33:22.923015  248918 custom.go:49] Setting up event handlers
I0922 13:33:22.923174  248918 custom.go:65] Starting controller
I0922 13:33:22.923199  248918 custom.go:66] Waiting for informer caches to sync
I0922 13:33:22.945836  248918 informer.go:66] Added object example
I0922 13:33:22.945869  248918 informer.go:71] Enqueue example to work queue
I0922 13:33:23.024154  248918 custom.go:71] Starting workers
I0922 13:33:23.024202  248918 custom.go:74] Started workers
I0922 13:33:23.024293  248918 reconciler.go:90] Hello world
```

# Running in Docker
```
$ kind create cluster
$ make apply-manifests
$ make build-image
$ make port-forward &
$ make push-image

$ kubectl --context=kind-kind get pods
NAME                          READY   STATUS    RESTARTS   AGE
controller-78bf6449cc-m8zqf   1/1     Running   0          4m12s
registry-0                    1/1     Running   0          4m12s

$ kubectl --context=kind-kind logs controller-78bf6449cc-m8zqf
I0921 22:42:22.981425       1 custom.go:49] Setting up event handlers
I0921 22:42:22.981752       1 custom.go:65] Starting controller
I0921 22:42:22.981780       1 custom.go:66] Waiting for informer caches to sync
I0921 22:42:23.088465       1 informer.go:66] Added object example
I0921 22:42:23.088501       1 informer.go:71] Enqueue example to work queue
I0921 22:42:23.182429       1 custom.go:71] Starting workers
I0921 22:42:23.182732       1 custom.go:74] Started workers
I0921 22:42:23.182991       1 reconciler.go:90] Hello world
```

# See also
* [sample-controller](https://github.com/kubernetes/sample-controller)
* [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder)
* [operator-sdk](https://github.com/operator-framework/operator-sdk)
* [kind](https://github.com/kubernetes-sigs/kind)
