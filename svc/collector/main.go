package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/livinlefevreloca/kantt/pkg/config"
	"github.com/livinlefevreloca/kantt/pkg/eventsource"
	"github.com/livinlefevreloca/kantt/pkg/storage"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"k8s.io/apimachinery/pkg/api/meta"
	api_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	events_v1 "k8s.io/client-go/kubernetes/typed/events/v1"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

func main() {
	// Setup database
	db := config.Database() // just a workaround, normally v should be passed via command line
	l := klog.Level(6)
	_ = l.Set("6")

	// get in cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic("Failed to get in cluster config: " + err.Error())
	}

	// Setup event watcher
	eventClient, err := events_v1.NewForConfig(config)
	if err != nil {
		panic("Failed to make eventCleint: " + err.Error())
	}
	eventWatcher, err := eventsource.NewWatcher(eventClient.Events(api_v1.NamespaceAll))
	if err != nil {
		panic("Failed to start eventWatcher: " + err.Error())
	}
	resultChan := eventWatcher.ResultChan()

	// Watch for SIGINT to shutdown gracefully
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	// Consume events forever or until SIGINT
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
	// Get or create Resource object
	resource := getOrCreateResource(db, event)
	// Create event object
	newEvent := storage.Event{
		Type:     event.Type,
		Resource: resource,
	}
	db.Create(&newEvent)
}

func getOrCreateResource(db *gorm.DB, event watch.Event) storage.Resource {
	obj := event.Object
	objMeta, err := meta.Accessor(obj)
	if err != nil {
		log.Println("error getting object meta: ", err)
	}
	// fmt.Println(obj)
	newResource := storage.Resource{
		GroupVersionKind: obj.GetObjectKind().GroupVersionKind(),
		Name:             objMeta.GetName(),
		Namespace:        objMeta.GetNamespace(),
	}
	result := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&newResource)
	if result.RowsAffected == 0 {
		// This will use the values present in newResource as a filter to try and
		// find the existing record in the database. If it finds one, it will
		// update newResource with the values from the database.
		db.First(&newResource)
	}
	return newResource
}
