package controller

import (
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	handlers "github.com/supercaracal/kubernetes-controller-template/internal/handler"
	workers "github.com/supercaracal/kubernetes-controller-template/internal/worker"
	clientset "github.com/supercaracal/kubernetes-controller-template/pkg/generated/clientset/versioned"
	customscheme "github.com/supercaracal/kubernetes-controller-template/pkg/generated/clientset/versioned/scheme"
	informers "github.com/supercaracal/kubernetes-controller-template/pkg/generated/informers/externalversions/supercaracal/v1"
	listers "github.com/supercaracal/kubernetes-controller-template/pkg/generated/listers/supercaracal/v1"
)

const (
	defaultReconcileDuration = 10 * time.Second
	componentName            = "kubernetes-controller-template"
	resourceName             = "FooBars"
)

// CustomController is
type CustomController struct {
	customClientSet      clientset.Interface
	jobSynced            cache.InformerSynced
	customResourceLister listers.FooBarLister
	customInformerSynced cache.InformerSynced
	workQueue            workqueue.RateLimitingInterface
	recorder             record.EventRecorder
	reconcileDuration    time.Duration
}

// NewCustomController is
func NewCustomController(kubeClientSet kubernetes.Interface, customClientSet clientset.Interface, customInformer informers.FooBarInformer) *CustomController {
	utilruntime.Must(customscheme.AddToScheme(scheme.Scheme))
	klog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartStructuredLogging(0)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeClientSet.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: componentName})
	wq := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), resourceName)

	controller := CustomController{
		customClientSet:      customClientSet,
		customResourceLister: customInformer.Lister(),
		customInformerSynced: customInformer.Informer().HasSynced,
		workQueue:            wq,
		recorder:             recorder,
		reconcileDuration:    defaultReconcileDuration,
	}

	klog.Info("Setting up event handlers")
	h := handlers.NewInformerHandler()
	customInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    h.OnAdd,
		UpdateFunc: h.OnUpdate,
		DeleteFunc: h.OnDelete,
	})

	return &controller
}

// Run is
func (c *CustomController) Run(stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workQueue.ShutDown()

	klog.Info("Starting controller")
	klog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.jobSynced, c.customInformerSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	klog.Info("Starting workers")
	rw := workers.NewReconciler(
		c.customClientSet,
		c.customResourceLister,
		c.workQueue,
		c.recorder,
	)

	go wait.Until(rw.Run, c.reconcileDuration, stopCh)

	klog.Info("Started workers")
	<-stopCh
	klog.Info("Shutting down workers")

	return nil
}
