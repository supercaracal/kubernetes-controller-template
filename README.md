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
I0923 10:25:05.032597  763146 informer.go:66] Added object example
I0923 10:25:05.032879  763146 informer.go:71] Enqueue example to work queue
I0923 10:25:05.105942  763146 custom.go:121] Controller is ready
I0923 10:25:05.106051  763146 reconciler.go:100] Hello world
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
I0922 22:53:20.279337       1 informer.go:66] Added object example
I0922 22:53:20.377106       1 informer.go:71] Enqueue example to work queue
I0922 22:53:20.976170       1 custom.go:121] Controller is ready
I0922 22:53:20.976253       1 reconciler.go:100] Hello world
```

## See also
* [sample-controller](https://github.com/kubernetes/sample-controller)
* [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder)
* [operator-sdk](https://github.com/operator-framework/operator-sdk)
* [kind](https://github.com/kubernetes-sigs/kind)
