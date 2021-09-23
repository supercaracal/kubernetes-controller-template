package worker

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	corelisterv1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	kubeerrors "k8s.io/cri-api/pkg/errors"
	"k8s.io/klog/v2"

	customapiv1 "github.com/supercaracal/kubernetes-controller-template/pkg/apis/supercaracal/v1"
	customclient "github.com/supercaracal/kubernetes-controller-template/pkg/generated/clientset/versioned"
	customlisterv1 "github.com/supercaracal/kubernetes-controller-template/pkg/generated/listers/supercaracal/v1"
)

// Reconciler is
type Reconciler struct {
	client    *ResourceClient
	lister    *ResourceLister
	workQueue workqueue.RateLimitingInterface
}

// ResourceClient is
type ResourceClient struct {
	Kube   kubernetes.Interface
	Custom customclient.Interface
}

// ResourceLister is
type ResourceLister struct {
	Pod            corelisterv1.PodLister
	CustomResource customlisterv1.FooBarLister
}

const (
	resourceName  = "FooBar"
	childLifetime = 5 * time.Minute
)

// NewReconciler is
func NewReconciler(cli *ResourceClient, list *ResourceLister, wq workqueue.RateLimitingInterface) *Reconciler {
	return &Reconciler{client: cli, lister: list, workQueue: wq}
}

// Run is
func (r *Reconciler) Run() {
	for r.processNextWorkItem() {
	}
}

func (r *Reconciler) processNextWorkItem() bool {
	obj, shutdown := r.workQueue.Get()
	if shutdown {
		return false
	}

	if err := r.sync(obj); err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}

func (r *Reconciler) sync(obj interface{}) error {
	defer r.workQueue.Done(obj)

	var key string
	var ok bool

	if key, ok = obj.(string); !ok {
		r.workQueue.Forget(obj)
		return fmt.Errorf("expected string in workqueue but got %#v", obj)
	}

	if err := r.create(key); err != nil {
		r.workQueue.AddRateLimited(key)
		return fmt.Errorf("error syncing '%s': %w, requeuing", key, err)
	}

	r.workQueue.Forget(obj)
	return nil
}

func (r *Reconciler) create(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return fmt.Errorf("invalid resource key: %s: %w", key, err)
	}

	parent, err := r.lister.CustomResource.FooBars(namespace).Get(name)
	if err != nil {
		if kubeerrors.IsNotFound(err) {
			return fmt.Errorf("custom resource '%s' in work queue no longer exists: %w", key, err)
		}

		return err
	}

	klog.Infof("Dequeued object %s successfully from work queue", key)
	child, err := r.createChildPod(parent)
	if err != nil {
		return err
	}
	klog.Infof("Created resource %s/%s successfully", child.Namespace, child.Name)

	return r.update(parent)
}

func (r *Reconciler) update(obj *customapiv1.FooBar) (err error) {
	cpy := obj.DeepCopy()
	cpy.Status.Succeeded = true
	_, err = r.client.Custom.SupercaracalV1().FooBars(obj.Namespace).Update(context.TODO(), cpy, metav1.UpdateOptions{})
	return
}

func (r *Reconciler) createChildPod(parent *customapiv1.FooBar) (*corev1.Pod, error) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%d", parent.Name, time.Now().Unix()),
			Namespace: parent.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(parent, customapiv1.SchemeGroupVersion.WithKind(resourceName)),
			},
		},
		Spec: corev1.PodSpec{
			RestartPolicy: corev1.RestartPolicyNever,
			Containers: []corev1.Container{
				corev1.Container{
					Name:    "main",
					Image:   "gcr.io/distroless/static-debian11:debug-amd64",
					Command: []string{"echo"},
					Args:    []string{parent.Spec.Message},
					SecurityContext: &corev1.SecurityContext{
						ReadOnlyRootFilesystem: func(b bool) *bool { return &b }(true),
					},
				},
			},
		},
	}

	return r.client.Kube.CoreV1().Pods(parent.Namespace).Create(context.TODO(), pod, metav1.CreateOptions{})
}

// Clean is
func (r *Reconciler) Clean() {
	parents, err := r.lister.CustomResource.List(labels.Everything())
	if err != nil {
		if !kubeerrors.IsNotFound(err) {
			utilruntime.HandleError(err)
		}
		return
	}

	background := metav1.DeletePropagationBackground
	baseTime := metav1.NewTime(time.Now().Add(-childLifetime))

	for _, parent := range parents {
		children, err := r.lister.Pod.Pods(parent.Namespace).List(labels.Everything())
		if err != nil {
			if !kubeerrors.IsNotFound(err) {
				utilruntime.HandleError(err)
			}
			continue
		}

		for _, child := range children {
			if !metav1.IsControlledBy(child, parent) {
				continue
			}

			if child.Status.Phase != corev1.PodSucceeded {
				continue
			}

			if baseTime.Before(child.Status.StartTime) || baseTime.Equal(child.Status.StartTime) {
				continue
			}

			if err := r.client.Kube.CoreV1().Pods(parent.Namespace).Delete(
				context.TODO(), child.Name, metav1.DeleteOptions{PropagationPolicy: &background},
			); err != nil {
				utilruntime.HandleError(err)
				continue
			}

			klog.Infof("Deleted resource %s/%s successfully", child.Namespace, child.Name)
		}
	}
}
