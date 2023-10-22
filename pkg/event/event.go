package event

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
)

type Notification struct {
	EventType     watch.EventType
	Endpoints     *corev1.Endpoints
	ClusterName   string
	Labels        map[string]string
	EndpointsName string
}

type Notifier interface {
	Send(notif *Notification)
	ReadChan() <-chan Notification
}

type Unbuffered struct {
	c chan Notification
}

func New() *Unbuffered {
	return &Unbuffered{
		c: make(chan Notification),
	}
}

func (n *Unbuffered) Send(notif *Notification) {
	n.c <- *notif
}

func (n *Unbuffered) ReadChan() <-chan Notification {
	return n.c
}
