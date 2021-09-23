![](https://github.com/supercaracal/kubernetes-controller-template/workflows/Test/badge.svg?branch=master)
![](https://github.com/supercaracal/kubernetes-controller-template/workflows/Release/badge.svg)

Kubernetes Controller Template
===============================================================================

This controller has a feature to log a message declared by manifest.

## Running on local host
```
$ kind create cluster
$ make apply-manifests
$ make build
$ make run
I0923 11:07:43.962836  779428 informer.go:66] Added object default/example
I0923 11:07:43.964894  779428 informer.go:71] Enqueue object default/example to work queue
I0923 11:07:44.040324  779428 custom.go:121] Controller is ready
I0923 11:07:44.040418  779428 reconciler.go:100] Dequeued object default/example successfully from work queue
I0923 11:07:44.040773  779428 reconciler.go:101] Hello world
```

## Running in Docker
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
I0923 11:24:12.777814       1 informer.go:66] Added object default/example
I0923 11:24:12.778102       1 informer.go:71] Enqueue object default/example to work queue
I0923 11:24:13.180382       1 custom.go:121] Controller is ready
I0923 11:24:13.180705       1 reconciler.go:100] Dequeued object default/example successfully from work queue
I0923 11:24:13.180791       1 reconciler.go:101] Hello world
```

## See also
* [sample-controller](https://github.com/kubernetes/sample-controller)
* [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder)
* [operator-sdk](https://github.com/operator-framework/operator-sdk)
* [kind](https://github.com/kubernetes-sigs/kind)
