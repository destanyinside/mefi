package localWatcher

import (
	"context"
	"github.com/destanyinside/mefi/pkg/event"
	"github.com/destanyinside/mefi/pkg/log"
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

type Watcher struct {
	sync.RWMutex
	logger                    *log.LogrusLogger
	localLabelSelector        string
	reSyncInt                 int
	mefiNamespace             string
	originalNameLabelSelector string
	client                    *structs.K8sClient
	event                     event.Notifier
	doneCh                    chan struct{}
	stopCh                    chan struct{}
	watcher                   *toolsWatch.RetryWatcher
}

func NewWatcher(logger *log.LogrusLogger, localLabelSelector string, originalNameLabelSelector string, reSyncInt int, mefiNamespace string, client *structs.K8sClient, event event.Notifier) *Watcher {
	return &Watcher{
		logger:                    logger,
		localLabelSelector:        localLabelSelector,
		originalNameLabelSelector: originalNameLabelSelector,
		reSyncInt:                 reSyncInt,
		mefiNamespace:             mefiNamespace,
		client:                    client,
		event:                     event,
	}
}

func (w *Watcher) Start() {
	w.logger.Infof("Starting all local watcher for %s", w.client.ClusterName)

	w.doneCh = make(chan struct{})
	w.stopCh = make(chan struct{})

	watchFunc := func(options metav1.ListOptions) (watch.Interface, error) {
		timeOut := int64(w.reSyncInt)
		return w.client.ClientSet.CoreV1().Endpoints(w.mefiNamespace).Watch(
			context.TODO(),
			metav1.ListOptions{
				TimeoutSeconds: &timeOut,
				LabelSelector:  w.localLabelSelector,
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
		endpointsByLabel, err := w.client.ClientSet.CoreV1().Endpoints(w.mefiNamespace).List(
			context.TODO(),
			metav1.ListOptions{
				LabelSelector: w.originalNameLabelSelector + "=" + item.Labels[w.originalNameLabelSelector],
			})
		if err != nil {
			w.logger.Errorf("Error list endpoints by label %s=%s", w.originalNameLabelSelector, item.Labels[w.originalNameLabelSelector])
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
				EventType:     eventType,
				Endpoints:     &newEndpoint,
				ClusterName:   w.client.ClusterName,
				EndpointsName: item.Labels[w.originalNameLabelSelector],
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
