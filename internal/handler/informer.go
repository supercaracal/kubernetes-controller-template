package handler

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

// InformerHandler is
type InformerHandler struct {
	wq workqueue.RateLimitingInterface
}

// NewInformerHandler is
func NewInformerHandler(wq workqueue.RateLimitingInterface) *InformerHandler {
	return &InformerHandler{wq: wq}
}

// OnAdd is
func (h *InformerHandler) OnAdd(obj interface{}) {
	h.handleObject(obj, "Added")
}

// OnUpdate is
func (h *InformerHandler) OnUpdate(before, after interface{}) {
	h.handleObject(after, "Updated")
}

// OnDelete is
func (h *InformerHandler) OnDelete(obj interface{}) {
	h.handleObject(obj, "Deleted")
}

func (h *InformerHandler) handleObject(obj interface{}, event string) {
	var object metav1.Object
	var ok bool

	if object, ok = obj.(metav1.Object); !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("error decoding object, invalid type"))
			return
		}

		object, ok = tombstone.Obj.(metav1.Object)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("error decoding object tombstone, invalid type"))
			return
		}

		klog.V(4).Infof("Recovered deleted object '%s' from tombstone", object.GetName())
	}

	klog.V(4).Infof("%s object %s", event, object.GetName())

	if event == "Deleted" {
		return
	}

	h.enqueueCustomResource(object)
}

func (h *InformerHandler) enqueueCustomResource(obj interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		utilruntime.HandleError(err)
		return
	}

	h.wq.Add(key)
}
