package localWatcher

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
	w.logger.Infof("Starting all local watcher for %s", w.client.ClusterName)

	w.doneCh = make(chan struct{})
	w.stopCh = make(chan struct{})

	//TODO() move namespace in config
	namespace := "mefi-system"

	watchFunc := func(options metav1.ListOptions) (watch.Interface, error) {
		// TODO() Set timeout in minutes from config
		timeOut := int64(0)
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
		var eventType watch.EventType
		// TODO() move labels and namespace in config
		endpointsByLabel, err := w.client.ClientSet.CoreV1().Endpoints("mefi-system").List(
			context.TODO(),
			metav1.ListOptions{
				LabelSelector: "originalName=" + item.Labels["originalName"],
			})
		if err != nil {
			// TODO() move labels in config
			w.logger.Errorf("Error list endpoints by label %s originalName=", item.Labels["originalName"])
		}
		if len(endpointsByLabel.Items) == 0 {
			eventType = watch.Deleted
		} else {
			eventType = watch.Added
		}
		var addresses, notReadyAddresses []corev1.EndpointAddress
		var ports []corev1.EndpointPort
		for _, subset := range endpointsByLabel.Items {
			for _, address := range subset.Subsets {
				addresses = append(addresses, address.Addresses...)
				notReadyAddresses = append(notReadyAddresses, address.NotReadyAddresses...)
				if len(ports) == 0 {
					ports = append(ports, address.Ports...)
				}
			}
		}

		newEndpoint := corev1.Endpoints{
			TypeMeta:   item.TypeMeta,
			ObjectMeta: item.ObjectMeta,
			Subsets: []corev1.EndpointSubset{
				{
					Addresses:         addresses,
					NotReadyAddresses: notReadyAddresses,
					Ports:             ports,
				},
			},
		}

		w.enqueue(
			&event.Notification{
				EventType:   eventType,
				Endpoints:   &newEndpoint,
				ClusterName: w.client.ClusterName,
				// TODO() move labels in config
				EndpointsName: item.Labels["originalName"],
			})
	}
}

func (w *Watcher) enqueue(event *event.Notification) {
	w.event.Send(event)
}

func (w *Watcher) Stop() {
	w.logger.Infof("Stopping all remoteWatcher for %s", w.client.ClusterName)
	w.stopCh <- struct{}{}
	w.RLock()
	w.RUnlock()
	<-w.doneCh
}
