package controller

import (
	"fmt"
	"time"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	handlers "github.com/supercaracal/kubernetes-controller-template/internal/handler"
	workers "github.com/supercaracal/kubernetes-controller-template/internal/worker"
	clientset "github.com/supercaracal/kubernetes-controller-template/pkg/generated/clientset/versioned"
	customscheme "github.com/supercaracal/kubernetes-controller-template/pkg/generated/clientset/versioned/scheme"
	informers "github.com/supercaracal/kubernetes-controller-template/pkg/generated/informers/externalversions"
	informerv1 "github.com/supercaracal/kubernetes-controller-template/pkg/generated/informers/externalversions/supercaracal/v1"
)

const (
	informerReSyncDuration = 5 * time.Second
	reconcileDuration      = 5 * time.Second
	resourceName           = "FooBars"
)

// CustomController is
type CustomController struct {
	customClientSet       clientset.Interface
	customInformerFactory informers.SharedInformerFactory
	customInformer        informerv1.FooBarInformer
	workQueue             workqueue.RateLimitingInterface
	reconcileDuration     time.Duration
}

// NewCustomController is
func NewCustomController(cfg *rest.Config) (*CustomController, error) {
	if err := customscheme.AddToScheme(scheme.Scheme); err != nil {
		return nil, err
	}

	cli, err := clientset.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	infoF := informers.NewSharedInformerFactory(cli, informerReSyncDuration)
	info := infoF.Supercaracal().V1().FooBars()

	wq := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), resourceName)
	h := handlers.NewInformerHandler(wq)
	info.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    h.OnAdd,
		UpdateFunc: h.OnUpdate,
		DeleteFunc: h.OnDelete,
	})

	controller := CustomController{
		customClientSet:       cli,
		customInformerFactory: infoF,
		customInformer:        info,
		workQueue:             wq,
		reconcileDuration:     reconcileDuration,
	}

	return &controller, nil
}

// Run is
func (c *CustomController) Run(stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workQueue.ShutDown()

	c.customInformerFactory.Start(stopCh)

	if ok := cache.WaitForCacheSync(stopCh, c.customInformer.Informer().HasSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	worker := workers.NewReconciler(c.customClientSet, c.customInformer.Lister(), c.workQueue)
	go wait.Until(worker.Run, c.reconcileDuration, stopCh)

	klog.Info("Controller is ready")
	<-stopCh
	klog.Info("Shutting down controller")

	return nil
}
