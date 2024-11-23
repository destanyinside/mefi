package localApplier

import (
	"context"
	"github.com/destanyinside/mefi/pkg/event"
	"github.com/destanyinside/mefi/pkg/log"
	"github.com/destanyinside/mefi/pkg/structs"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"sync"
	"time"
)

type Applier struct {
	sync.RWMutex
	logger        *log.LogrusLogger
	mefiNamespace string
	client        *structs.K8sClient
	event         event.Notifier
	doneCh        chan struct{}
	stopCh        chan struct{}
}

func NewApplier(logger *log.LogrusLogger, mefiNamespace string, client *structs.K8sClient, event event.Notifier) *Applier {
	return &Applier{
		logger:        logger,
		mefiNamespace: mefiNamespace,
		client:        client,
		event:         event,
	}
}

func (a *Applier) Start() {
	a.logger.Infof("Starting kubernetes localApplier endpoints")

	a.doneCh = make(chan struct{})
	a.stopCh = make(chan struct{})

	go func() {
		ticker := time.NewTicker(180 * time.Second)
		defer ticker.Stop()
		defer close(a.doneCh)
		for {
			evChan := a.event.ReadChan()
			select {
			case <-a.stopCh:
				return
			case ev := <-evChan:
				// TODO() this place may be wrong or something like this
				go func() {
					a.processEvent(&ev)
				}()
			case <-ticker.C:
			}
		}
	}()
}

func (a *Applier) processEvent(ev *event.Notification) {
	switch ev.EventType {
	case watch.Added:
		a.create(ev.ClusterName, ev.Endpoints, ev.EndpointsName, ev.Labels)
	case watch.Modified:
		a.update(ev.ClusterName, ev.Endpoints, ev.EndpointsName, ev.Labels)
	case watch.Deleted:
		a.delete(ev.ClusterName, ev.EndpointsName)
	case watch.Bookmark:
		a.logger.Infof("Receive update only ResourceVersion for %s from cluster %s. Skip", ev.EndpointsName, ev.ClusterName)
	case watch.Error:
		a.logger.Errorf("Receive error event for endpoint %s from cluster %s", ev.EndpointsName, ev.ClusterName)
	}
}

func (a *Applier) create(clusterName string, object *corev1.Endpoints, endpointsName string, labels map[string]string) {
	newEndpoint := &corev1.Endpoints{
		TypeMeta: object.TypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:   endpointsName,
			Labels: labels,
		},
		Subsets: object.Subsets,
	}

	_, err := a.client.ClientSet.CoreV1().Endpoints(a.mefiNamespace).Create(
		context.TODO(),
		newEndpoint,
		metav1.CreateOptions{})
	switch {
	case apierrors.IsAlreadyExists(err):
		a.update(clusterName, object, endpointsName, labels)
	case err == nil:
		a.logger.Infof("Create endpoint %s from cluster %s", endpointsName, clusterName)
	default:
		a.logger.Errorf("Error create endpoint %s from cluster %s: %s", endpointsName, clusterName, err)
	}
}

func (a *Applier) update(clusterName string, object *corev1.Endpoints, endpointsName string, labels map[string]string) {
	newEndpoint := &corev1.Endpoints{
		TypeMeta: object.TypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:   endpointsName,
			Labels: labels,
		},
		Subsets: object.Subsets,
	}

	if _, err := a.client.ClientSet.CoreV1().Endpoints(a.mefiNamespace).Update(
		context.TODO(),
		newEndpoint,
		metav1.UpdateOptions{}); err != nil {
		a.logger.Errorf("Error update endpoint %s from cluster %s: %s", endpointsName, clusterName, err)
	} else {
		a.logger.Infof("Update endpoint %s from cluster %s", endpointsName, clusterName)
	}
}

func (a *Applier) delete(clusterName string, endpointsName string) {
	if err := a.client.ClientSet.CoreV1().Endpoints(a.mefiNamespace).Delete(
		context.TODO(),
		endpointsName,
		metav1.DeleteOptions{}); err != nil {
		a.logger.Errorf("Error delete endpoint %s from cluster %s: %s", endpointsName, clusterName, err)
	} else {
		a.logger.Infof("Delete endpoint %s from cluster %s", endpointsName, clusterName)
	}
}

func (a *Applier) Stop() {
	a.logger.Infof("Stopping localApplier for %s", a.client.ClusterName)
	a.stopCh <- struct{}{}
	a.RLock()
	a.RUnlock()
	<-a.doneCh
}
