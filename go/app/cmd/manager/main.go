package main

import (
	"disposableOperator/pkg/controller/podrestart"
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

const CUSTOM_NAMESPACE = "default" // Replace with your custom namespace
func main() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.InfoLevel)

	fmt.Println("Disposable Operator is starting up...")

	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("Failed to get in-cluster config: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	watchlist := cache.NewSharedInformer(
		cache.NewListWatchFromClient(
			clientset.CoreV1().RESTClient(),
			"pods",
			v1.NamespaceAll,
			fields.Everything(),
		),
		&v1.Pod{},
		0*time.Second,
	)
	watchlist.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj, newObj interface{}) {
			newPod := newObj.(*v1.Pod)
			oldPod := oldObj.(*v1.Pod)

			if strings.HasPrefix(newPod.ObjectMeta.Namespace, CUSTOM_NAMESPACE) {
				//				log.Infof("Detected a pod in namespace starting with 'CUSTOM_NAMESPACE': %s", newPod.ObjectMeta.Name)
				if podrestart.IsPodRestarting(oldPod, newPod) {
					log.Infof("Pod %s appears to be restarting. Attempting to handle...", newPod.ObjectMeta.Name)
					err := podrestart.HandlePodRestart(clientset, newPod)
					if err != nil {
						log.Errorf("Error handling pod restart for pod %s: %v", newPod.ObjectMeta.Name, err)
					}
				}
			}
		},
	})
	fmt.Println("Operator is now watching for pod updates :) ")

	stop := make(chan struct{})
	defer close(stop)

	log.Info("Starting watchlist...")
	go watchlist.Run(stop)
	log.Info("Watchlist started.")

	select {}
}
