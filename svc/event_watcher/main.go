package main

import (
	"github.com/endophage/kant/lib/config"
	"github.com/endophage/kant/lib/eventsource"
	"gorm.io/gorm"
	"k8s.io/apimachinery/pkg/watch"
	v1 "k8s.io/client-go/kubernetes/typed/events/v1"
	"k8s.io/client-go/rest"
	"os"
	"os/signal"
)

func main() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	db := config.Database()
	eventClient, err := v1.NewForConfig(&rest.Config{})
	if err != nil {
		panic(err)
	}
	eventWatcher, err := eventsource.NewWatcher(eventClient.Events(""))
	resultChan := eventWatcher.ResultChan()
	if err != nil {
		panic(err)
	}
	for {
		select {
		case event := <-resultChan:
			processEvent(db, event)
		case <-signalChan:
			eventWatcher.Stop()
			goto exit
		}
	}

exit:
	return
}

func processEvent(db *gorm.DB, event watch.Event) {
	switch event.Type {
	case watch.Added:
		db.Create(&event.Object)
	case watch.Modified:
		db.Save(&event.Object)
	case watch.Deleted:
		db.Delete(&event.Object)
	}
}
