package discovery

import (
	"github.com/destanyinside/mefi/pkg/controller"
	"github.com/destanyinside/mefi/pkg/structs"
	"golang.org/x/net/context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
	"sync"
	"time"
)

type Discover struct {
	sync.RWMutex
	logger    logger
	factory   ControllerFactory
	ctrls     controllerCollection
	selector  string
	doneCh    chan struct{}
	stopCh    chan struct{}
	k8sClient *structs.K8sClient
}

type controllerCollection map[string]controller.Interface

type logger interface {
	Infof(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

type ControllerFactory interface {
	NewController(client cache.ListerWatcher, name string, clusterName string) controller.Interface
}

func New(log logger, factory ControllerFactory, selector string, k8sClient *structs.K8sClient) *Discover {
	return &Discover{
		logger:    log,
		factory:   factory,
		ctrls:     make(controllerCollection),
		selector:  selector,
		k8sClient: k8sClient,
	}
}

func (d *Discover) Start() *Discover {
	d.logger.Infof("Starting all kubernetes controllers")

	d.doneCh = make(chan struct{})
	d.stopCh = make(chan struct{})

	go func() {
		ticker := time.NewTicker(180 * time.Second)
		defer ticker.Stop()
		defer close(d.doneCh)

		for {
			err := d.discovery()
			if err != nil {
				d.logger.Errorf("Refresh failed: %v", err)
			}

			select {
			case <-d.stopCh:
				return
			case <-ticker.C:
			}
		}
	}()

	return d
}

func (d *Discover) discovery() error {

	d.Lock()
	defer d.Unlock()

	namespace := metav1.NamespaceAll

	resource := schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "endpoints",
	}

	kind := "endpoints"

	lw := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			return d.k8sClient.DInf.Resource(resource).Namespace(namespace).List(context.TODO(), options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return d.k8sClient.DInf.Resource(resource).Namespace(namespace).Watch(context.TODO(), options)
		},
	}

	d.ctrls[d.k8sClient.ClusterName] = d.factory.NewController(lw, kind, d.k8sClient.ClusterName)
	go d.ctrls[d.k8sClient.ClusterName].Start()

	return nil
}

func (d *Discover) Stop() {
	d.logger.Infof("Stopping all kubernetes controllers")
	d.stopCh <- struct{}{}
	d.RLock()
	d.RUnlock()
	<-d.doneCh
}
