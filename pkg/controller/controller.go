package controller

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"strings"
	"time"
)

var (
	maxProcessRetry = 6
	canaryKey       = "$mefi canary$"
)

type logger interface {
	Infof(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

type Factory struct {
	logger     logger
	selector   string
	resyncIntv time.Duration
}

type Interface interface {
	Start()
	Stop()
}

type Controller struct {
	name        string
	stopCh      chan struct{}
	doneCh      chan struct{}
	syncCh      chan struct{}
	queue       workqueue.RateLimitingInterface
	informer    cache.SharedIndexInformer
	logger      logger
	resyncIntv  time.Duration
	clusterName string
}

// NewFactory create a controller factory
func NewFactory(logger logger, selector string, resync int) *Factory {
	return &Factory{
		logger:     logger,
		selector:   selector,
		resyncIntv: time.Duration(resync) * time.Second,
	}
}

func New(client cache.ListerWatcher,
	log logger,
	name string,
	selector string,
	resync time.Duration,
	clusterName string,
) *Controller {

	lopts := metav1.ListOptions{LabelSelector: selector, ResourceVersion: "0", AllowWatchBookmarks: true}
	lw := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			return client.List(lopts)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return client.Watch(lopts)
		},
	}

	informer := cache.NewSharedIndexInformer(
		lw,
		&unstructured.Unstructured{},
		resync,
		cache.Indexers{},
	)

	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(key)
			}
		},
		UpdateFunc: func(old, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err == nil {
				queue.Add(key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(key)
			}
		},
	})

	return &Controller{
		stopCh:      make(chan struct{}),
		doneCh:      make(chan struct{}),
		syncCh:      make(chan struct{}, 1),
		name:        name,
		queue:       queue,
		informer:    informer,
		logger:      log,
		resyncIntv:  resync,
		clusterName: clusterName,
	}
}

func (c *Controller) Start() {
	c.logger.Infof("Starting %s controller for %s cluster", c.name, c.clusterName)
	defer utilruntime.HandleCrash()

	go c.informer.Run(c.stopCh)

	if !cache.WaitForCacheSync(c.stopCh, c.informer.HasSynced) {
		utilruntime.HandleError(fmt.Errorf("timed out waiting for cache sync"))
		return
	}

	c.queue.Add(canaryKey)

	go wait.Until(c.runWorker, time.Second, c.stopCh)
}

// Stop halts the controller
func (c *Controller) Stop() {
	c.logger.Infof("Stopping %s controller for %s cluster", c.name, c.clusterName)
	<-c.syncCh
	close(c.stopCh)
	c.queue.ShutDown()
	<-c.doneCh
}

func (c *Controller) runWorker() {
	defer close(c.doneCh)
	for c.processNextItem() {
		// continue looping
	}
}

func (c *Controller) processNextItem() bool {
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	defer c.queue.Done(key)

	if strings.Compare(key.(string), canaryKey) == 0 {
		c.logger.Infof("Initial sync completed for %s controller", c.clusterName)
		c.syncCh <- struct{}{}
		c.queue.Forget(key)
		return true
	}

	err := c.processItem(key.(string))

	if err == nil {
		// No error, reset the ratelimit counters
		c.queue.Forget(key)
	} else if c.queue.NumRequeues(key) < maxProcessRetry {
		c.logger.Errorf("Error processing %s (will retry): %v", key, err)
		c.queue.AddRateLimited(key)
	} else {
		// err != nil and too many retries
		c.logger.Errorf("Error processing %s (giving up): %v", key, err)
		c.queue.Forget(key)
	}

	return true
}

func (c *Controller) processItem(key string) error {
	rawobj, exists, err := c.informer.GetIndexer().GetByKey(key)
	if err != nil {
		return fmt.Errorf("error fetching %s from store: %v", key, err)
	}

	if !exists {
		//TODO() send notify with delete endpoint
		c.logger.Infof("Delete endpoint: %s", key)
		return nil
	}

	obj := rawobj.(*unstructured.Unstructured).DeepCopy()
	uc := obj.UnstructuredContent()
	//host := uc["metadata"].(map[string]interface{})["labels"].(map[string]interface{})["gec.io/host"]
	subsets := uc["subsets"]
	name := uc["metadata"].(map[string]interface{})["name"]
	c.logger.Infof("name: %s", name)
	c.logger.Infof("subsets: %v", subsets)
	c.logger.Infof("cluster: %s ", c.clusterName)

	return err
}

func (f *Factory) NewController(client cache.ListerWatcher, name string, clusterName string) Interface {
	return New(client, f.logger, name, f.selector, f.resyncIntv, clusterName)
}
