package worker

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/cri-api/pkg/errors"
	"k8s.io/klog/v2"

	customapiv1 "github.com/supercaracal/kubernetes-controller-template/pkg/apis/supercaracal/v1"
	clientset "github.com/supercaracal/kubernetes-controller-template/pkg/generated/clientset/versioned"
	listers "github.com/supercaracal/kubernetes-controller-template/pkg/generated/listers/supercaracal/v1"
)

// Reconciler is
type Reconciler struct {
	customClientSet      clientset.Interface
	customResourceLister listers.FooBarLister
	workQueue            workqueue.RateLimitingInterface
}

// NewReconciler is
func NewReconciler(customClientSet clientset.Interface, customResourceLister listers.FooBarLister, workQueue workqueue.RateLimitingInterface) *Reconciler {
	return &Reconciler{
		customClientSet:      customClientSet,
		customResourceLister: customResourceLister,
		workQueue:            workQueue,
	}
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

	if err := r.exec(obj); err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}

func (r *Reconciler) exec(obj interface{}) error {
	defer r.workQueue.Done(obj)

	var key string
	var ok bool

	if key, ok = obj.(string); !ok {
		r.workQueue.Forget(obj)
		return fmt.Errorf("expected string in workqueue but got %#v", obj)
	}

	if err := r.do(key); err != nil {
		r.workQueue.AddRateLimited(key)
		return fmt.Errorf("error syncing '%s': %w, requeuing", key, err)
	}

	r.workQueue.Forget(obj)
	return nil
}

func (r *Reconciler) do(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return fmt.Errorf("invalid resource key: %s: %w", key, err)
	}

	resource, err := r.customResourceLister.FooBars(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			return fmt.Errorf("custom resource '%s' in work queue no longer exists: %w", key, err)
		}

		return err
	}

	klog.Info(resource.Spec.Message)
	return r.updateCustomResourceStatus(resource)
}

func (r *Reconciler) updateCustomResourceStatus(resource *customapiv1.FooBar) (err error) {
	cpy := resource.DeepCopy()
	cpy.Status.Succeeded = true
	_, err = r.customClientSet.SupercaracalV1().FooBars(resource.Namespace).Update(context.TODO(), cpy, metav1.UpdateOptions{})
	return
}
