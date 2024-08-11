package main

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/livinlefevreloca/kantt/pkg/config"
	"github.com/livinlefevreloca/kantt/pkg/storage"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

func main() {
	// Setup database
	l := klog.Level(6)
	_ = l.Set("6")

	var db = config.Database() // just a workaround, normally v should be passed via command line
	// get in cluster config

	var local bool
	flag.BoolVar(&local, "local", true, "Use local config")
	flag.Parse()

	var config *rest.Config
	var err error
	if local {
		config, err = clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
		if err != nil {
			panic("Failed to get local config: " + err.Error())
		}
	} else {
		config, err = rest.InClusterConfig()
		if err != nil {
			panic("Failed to get in cluster config: " + err.Error())
		}
	}

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic("Failed to create client set: " + err.Error())
	}

	factory := informers.NewSharedInformerFactoryWithOptions(
		clientSet,
		0,
	)

	podInformer := factory.Core().V1().Pods().Informer()
	defer runtime.HandleCrash()

	stopChan := make(chan struct{})
	defer close(stopChan)

	go factory.Start(stopChan)

	// Watch for SIGINT to shutdown gracefully
	signalChan := make(chan os.Signal, 1)

	if !cache.WaitForCacheSync(stopChan, podInformer.HasSynced) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}

	addOrDeletePod := func(obj interface{}) {
		pod := obj.(*v1.Pod)
		slog.Info(
			"Pod event",
			"pod", pod.Name,
			"namespace", pod.Namespace,
			"phase", pod.Status.Phase,
		)
		createOrUpdatePod(db, pod)
	}

	// Handle add, update and delete events for pods and update the database
	// accordingly
	podInformer.AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: addOrDeletePod,
			UpdateFunc: func(oldObj, newObj interface{}) {
				oldPod := oldObj.(*v1.Pod)
				newPod := newObj.(*v1.Pod)
				slog.Info(
					"Update Pod event",
					"pod", newPod.Name,
					"namespace", newPod.Namespace,
					"phase", newPod.Status.Phase,
					"old_pod_name", oldPod.Name,
					"old_pod_namespace", oldPod.Namespace,
					"old_pod_phase", oldPod.Status.Phase,
				)
				createOrUpdatePod(db, newPod)
			},
			DeleteFunc: addOrDeletePod,
		},
	)

	select {
	case <-signalChan:
		klog.Info("Received SIGINT, shutting down")
	case <-stopChan:
		klog.Info("Received stop signal, shutting down")
	}
	return
}

// Update the database with the pod information as events are received
func createOrUpdatePod(db *gorm.DB, pod *v1.Pod) error {
	// Get the owner of the pod
	owner, err := GetOwner(db, pod)
	if err != nil {
		return err
	}
	db.Clauses(clause.OnConflict{DoNothing: true}).Create(owner)

	var storedPod *storage.Pod
	switch pod.Status.Phase {
	case v1.PodPending:
		storedPod = &storage.Pod{
			Name:        pod.Name,
			Namespace:   pod.Namespace,
			Owner:       *owner,
			PendingTime: time.Now(),
		}
		db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "name"}, {Name: "namespace"}},
			DoUpdates: clause.AssignmentColumns([]string{"pending_time"}),
		}).Create(storedPod)
	case v1.PodRunning:
		storedPod = &storage.Pod{
			Name:         pod.Name,
			Namespace:    pod.Namespace,
			Owner:        *owner,
			StartingTime: time.Now(),
		}
		db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "name"}, {Name: "namespace"}},
			DoUpdates: clause.AssignmentColumns([]string{"starting_time"}),
		}).Create(storedPod)
	case v1.PodSucceeded, v1.PodFailed:
		storedPod = &storage.Pod{
			Name:       pod.Name,
			Namespace:  pod.Namespace,
			Owner:      *owner,
			EndingTime: time.Now(),
		}
		db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "name"}, {Name: "namespace"}},
			DoUpdates: clause.AssignmentColumns([]string{"ending_time"}),
		}).Create(storedPod)
	case v1.PodUnknown:
		slog.Error(
			"Got pod in unknown phase ignoring",
			"pod", pod.Name,
			"namespace", pod.Namespace,
			"phase", pod.Status.Phase,
		)
	}

	// Store the node and the pod node relationship
	node, ok := createNode(db, pod)
	if ok {
		oldNodePods := []storage.NodePod{}
		db.Table("node_pods").Where("pod_id = ? AND end_time IS NULL", storedPod.ID).Find(&oldNodePods)

		if len(oldNodePods) == 0 {
			nodePod := storage.NodePod{
				Node:      *node,
				Pod:       *storedPod,
				StartTime: time.Now(),
			}
			db.Create(&nodePod)
		} else if len(oldNodePods) == 1 {
			oldNodePod := oldNodePods[0]
			db.Model(&oldNodePod).Update("end_time", time.Now())
			nodePod := storage.NodePod{
				Node:      *node,
				Pod:       *storedPod,
				StartTime: time.Now(),
			}
			db.Create(&nodePod)
		} else {
			slog.Error(
				"Found multiple active node pods for pod. Correcting",
				"pod", storedPod.Name,
				"namespace", storedPod.Namespace,
				"node", node.Name,
			)

			// Close the time window on all the old node pods
			for i := 0; i < len(oldNodePods); i++ {
				db.Model(&oldNodePods[i]).Update("end_time", time.Now())
			}

			nodePod := storage.NodePod{
				Node:      *node,
				Pod:       *storedPod,
				StartTime: time.Now(),
			}
			db.Create(&nodePod)
		}
	}

	return nil
}

const (
	OWNER_DEPLOYMENT  = "Deployment"
	OWNER_STATEFULSET = "StatefulSet"
	OWNER_DAEMONSET   = "DaemonSet"
)

func GetOwner(db *gorm.DB, pod *v1.Pod) (*storage.Owner, error) {
	var ownerTypes []string
	for _, ownerRef := range pod.OwnerReferences {
		ok, ownerKind := getOwnerKind(ownerRef.Kind)
		if !ok {
			ownerTypes = append(ownerTypes, ownerRef.Kind)
			continue
		}
		owner := storage.Owner{
			Name:      ownerRef.Name,
			Namespace: pod.Namespace,
			Kind:      ownerKind,
		}
		switch owner.Kind {
		case OWNER_DEPLOYMENT:
		case OWNER_STATEFULSET:
		case OWNER_DAEMONSET:
			return &owner, nil
		default:
			ownerTypes = append(ownerTypes, owner.Kind)
		}
	}
	return nil, errors.New(fmt.Sprintf("No Known owner type found for pod. Found: %v", ownerTypes))
}

func getOwnerKind(ownerType string) (bool, string) {
	switch ownerType {
	case "ReplicaSet":
		return true, OWNER_DEPLOYMENT
	case "StatefulSet":
		return true, OWNER_STATEFULSET
	case "DaemonSet":
		return true, OWNER_DAEMONSET
	default:
		return false, ""
	}
}

func createNode(db *gorm.DB, pod *v1.Pod) (*storage.Node, bool) {
	// Get the node information
	nodeName := pod.Spec.NodeName
	if nodeName == "" {
		return nil, false
	}
	nodeIP := pod.Status.HostIP

	node := storage.Node{
		Name: nodeName,
		IP:   nodeIP,
	}
	db.Clauses(clause.OnConflict{DoNothing: true}).Create(&node)
	return &node, true
}
