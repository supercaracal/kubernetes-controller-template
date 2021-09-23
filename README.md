![](https://github.com/supercaracal/kubernetes-controller-template/workflows/Test/badge.svg?branch=master)
![](https://github.com/supercaracal/kubernetes-controller-template/workflows/Release/badge.svg)

Kubernetes Controller Template
===============================================================================

This controller has a feature to create a pod to log a message declared by manifest.

```
$ kubectl --context=kind-kind get pods
NAME                          READY   STATUS      RESTARTS   AGE
controller-78bf6449cc-ptwnn   1/1     Running     0          17s
example-1632371659            0/1     Completed   0          14s
registry-0                    1/1     Running     0          15h

$ kubectl --context=kind-kind logs controller-78bf6449cc-ptwnn
I0923 13:34:18.976502       1 informer.go:66] Added object default/example
I0923 13:34:18.976700       1 informer.go:71] Enqueue object default/example to work queue
I0923 13:34:19.377067       1 custom.go:121] Controller is ready
I0923 13:34:19.377173       1 reconciler.go:106] Dequeued object default/example successfully from work queue
I0923 13:34:19.485510       1 reconciler.go:112] Created resource default/example-1632371659 successfully

$ kubectl --context=kind-kind logs example-1632371659
Hello world
```

## Running on local host
```
$ kind create cluster
$ make apply-manifests
$ make build
$ make run
```

## Running in Docker
```
$ kind create cluster
$ make apply-manifests
$ make build-image
$ make port-forward &
$ make push-image
```

## See also
* [sample-controller](https://github.com/kubernetes/sample-controller)
* [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder)
* [operator-sdk](https://github.com/operator-framework/operator-sdk)
* [kind](https://github.com/kubernetes-sigs/kind)

## TODO
You can edit the following files as needed.

```
$ grep -riIl --exclude-dir=generated --exclude-dir=.git --exclude=zz_generated.deepcopy.go 'supercaracal\|foobar\|kubernetes-controller-template' .
./README.md
./go.mod
./.github/workflows/release.yaml
./internal/controller/custom.go
./internal/worker/reconciler.go
./Makefile
./.dockerignore
./.gitignore
./config/controller.yaml
./config/registry.yaml
./config/crd.yaml
./config/example-foobar.yaml
./main.go
./pkg/apis/supercaracal/register.go
./pkg/apis/supercaracal/v1/doc.go
./pkg/apis/supercaracal/v1/register.go
./pkg/apis/supercaracal/v1/types.go
./Dockerfile
```
