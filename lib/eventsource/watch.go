package eventsource

import (
	"context"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/typed/events/v1"
)

// NewWatcher for kubernetes events and send them to the channel.
// This function is blocking.
func NewWatcher(eventsClient v1.EventInterface) (watch.Interface, error) {
	return eventsClient.Watch(context.Background(), meta.ListOptions{})
}
