package controller

import (
	"fmt"
	"time"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	kubescheme "k8s.io/client-go/kubernetes/scheme"
	corelisterv1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	handlers "github.com/supercaracal/kubernetes-controller-template/internal/handler"
	workers "github.com/supercaracal/kubernetes-controller-template/internal/worker"
	customclient "github.com/supercaracal/kubernetes-controller-template/pkg/generated/clientset/versioned"
	customscheme "github.com/supercaracal/kubernetes-controller-template/pkg/generated/clientset/versioned/scheme"
	custominformers "github.com/supercaracal/kubernetes-controller-template/pkg/generated/informers/externalversions"
	customlisterv1 "github.com/supercaracal/kubernetes-controller-template/pkg/generated/listers/supercaracal/v1"
)

const (
	informerReSyncDuration = 5 * time.Second
	reconcileDuration      = 5 * time.Second
	cleanupDuration        = 1 * time.Minute
	resourceName           = "FooBars"
)

// CustomController is
type CustomController struct {
	kube      *kubeTool
	custom    *customTool
	workQueue workqueue.RateLimitingInterface
}

type kubeTool struct {
	client  kubernetes.Interface
	factory kubeinformers.SharedInformerFactory
	pod     *podInfo
}

type customTool struct {
	client   customclient.Interface
	factory  custominformers.SharedInformerFactory
	resource *customResourceInfo
}

type podInfo struct {
	informer cache.SharedIndexInformer
	lister   corelisterv1.PodLister
}

type customResourceInfo struct {
	informer cache.SharedIndexInformer
	lister   customlisterv1.FooBarLister
}

// NewCustomController is
func NewCustomController(cfg *rest.Config) (*CustomController, error) {
	if err := customscheme.AddToScheme(kubescheme.Scheme); err != nil {
		return nil, err
	}

	kube, err := buildKubeTools(cfg)
	if err != nil {
		return nil, err
	}

	custom, err := buildCustomResourceTools(cfg)
	if err != nil {
		return nil, err
	}

	wq := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), resourceName)
	h := handlers.NewInformerHandler(wq)
	custom.resource.informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    h.OnAdd,
		UpdateFunc: h.OnUpdate,
		DeleteFunc: h.OnDelete,
	})

	return &CustomController{kube: kube, custom: custom, workQueue: wq}, nil
}

// Run is
func (c *CustomController) Run(stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workQueue.ShutDown()

	c.kube.factory.Start(stopCh)
	c.custom.factory.Start(stopCh)

	if ok := cache.WaitForCacheSync(stopCh, c.kube.pod.informer.HasSynced, c.custom.resource.informer.HasSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	worker := workers.NewReconciler(
		&workers.ResourceClient{
			Kube:   c.kube.client,
			Custom: c.custom.client,
		},
		&workers.ResourceLister{
			Pod:            c.kube.pod.lister,
			CustomResource: c.custom.resource.lister,
		},
		c.workQueue,
	)
	go wait.Until(worker.Run, reconcileDuration, stopCh)
	go wait.Until(worker.Clean, cleanupDuration, stopCh)

	klog.Info("Controller is ready")
	<-stopCh
	klog.Info("Shutting down controller")

	return nil
}

func buildKubeTools(cfg *rest.Config) (*kubeTool, error) {
	cli, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	info := kubeinformers.NewSharedInformerFactory(cli, informerReSyncDuration)
	pod := info.Core().V1().Pods()

	return &kubeTool{
		client:  cli,
		factory: info,
		pod:     &podInfo{informer: pod.Informer(), lister: pod.Lister()},
	}, nil
}

func buildCustomResourceTools(cfg *rest.Config) (*customTool, error) {
	cli, err := customclient.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	info := custominformers.NewSharedInformerFactory(cli, informerReSyncDuration)
	cr := info.Supercaracal().V1().FooBars()

	return &customTool{
		client:   cli,
		factory:  info,
		resource: &customResourceInfo{informer: cr.Informer(), lister: cr.Lister()},
	}, nil
}
