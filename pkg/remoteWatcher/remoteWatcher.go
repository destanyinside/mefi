package remoteWatcher

import (
	"context"
	"github.com/destanyinside/mefi/pkg/event"
	"github.com/destanyinside/mefi/pkg/structs"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
	toolsWatch "k8s.io/client-go/tools/watch"
	"sync"
	"time"
)

type logger interface {
	Infof(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

type Watcher struct {
	sync.RWMutex
	logger   logger
	selector string
	client   *structs.K8sClient
	event    event.Notifier
	doneCh   chan struct{}
	stopCh   chan struct{}
	watcher  *toolsWatch.RetryWatcher
}

func NewWatcher(logger logger, selector string, client *structs.K8sClient, event event.Notifier) *Watcher {
	return &Watcher{
		logger:   logger,
		selector: selector,
		client:   client,
		event:    event,
	}
}

func (w *Watcher) Start() {
	w.logger.Infof("Starting all remote watcher for %s", w.client.ClusterName)

	w.doneCh = make(chan struct{})
	w.stopCh = make(chan struct{})

	namespace := metav1.NamespaceAll

	watchFunc := func(options metav1.ListOptions) (watch.Interface, error) {
		timeOut := int64(60)
		return w.client.ClientSet.CoreV1().Endpoints(namespace).Watch(
			context.TODO(),
			metav1.ListOptions{
				TimeoutSeconds: &timeOut,
				LabelSelector:  w.selector,
			})
	}

	w.watcher, _ = toolsWatch.NewRetryWatcher("1", &cache.ListWatch{WatchFunc: watchFunc})

	go func() {
		ticker := time.NewTicker(180 * time.Second)
		defer ticker.Stop()
		defer close(w.doneCh)

		for {
			go wait.Until(w.runWorker, time.Second, w.stopCh)

			select {
			case <-w.stopCh:
				return
			case <-ticker.C:
			}
		}
	}()
}

func (w *Watcher) runWorker() {
	defer close(w.doneCh)
	for endpoints := range w.watcher.ResultChan() {
		item := endpoints.Object.(*corev1.Endpoints)
		labels := make(map[string]string)
		// TODO() move labels in config
		labels["isMefiExported"] = "true"
		labels["originalName"] = item.Name
		w.enqueue(
			&event.Notification{
				EventType:     endpoints.Type,
				Endpoints:     item,
				ClusterName:   w.client.ClusterName,
				Labels:        labels,
				EndpointsName: item.Name + "-" + w.client.ClusterName,
			})
	}
}

func (w *Watcher) enqueue(event *event.Notification) {
	w.event.Send(event)
}

func (w *Watcher) Stop() {
	w.logger.Infof("Stopping all remote watcher for %s", w.client.ClusterName)
	w.stopCh <- struct{}{}
	w.RLock()
	w.RUnlock()
	<-w.doneCh
}
