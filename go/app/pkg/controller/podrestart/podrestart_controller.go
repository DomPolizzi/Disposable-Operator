package podrestart

import (
	"context"
	"fmt"
	"log"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const RESTART_THRESHOLD = 3 // Number of restarts before scaling down the deployment
const MAX_RETRIES = 3       // Amount of times the operator will retry handling a pod restart

func IsPodRestarting(oldPod, newPod *v1.Pod) bool {
	log.Printf("Checking State of Pod %s in %s \n", newPod.Name, newPod.Namespace)

	for _, newContainerStatus := range newPod.Status.ContainerStatuses {
		for _, oldContainerStatus := range oldPod.Status.ContainerStatuses {
			if oldContainerStatus.Name == newContainerStatus.Name {
				if newContainerStatus.RestartCount > oldContainerStatus.RestartCount+RESTART_THRESHOLD {
					return true
				}

				// Check State
				if newContainerStatus.State.Waiting != nil && (newContainerStatus.State.Waiting.Reason == "CrashLoopBackOff" || newContainerStatus.State.Waiting.Reason == "Error" || newContainerStatus.State.Waiting.Reason == "ContainerStatusUnknown") {
					log.Printf("Pod %s in %s is in Bad State... \n", newPod.Name, newPod.Namespace)
					return true
				}
			}
		}
	}
	return false
}

func HandlePodRestart(clientset *kubernetes.Clientset, pod *v1.Pod) error {
	log.Printf("Handling Pod %s in %s \n", pod.Name, pod.Namespace)

	deploymentName := pod.Labels["app"]
	if deploymentName == "" {
		log.Printf("No Deployment Found for %s in %s \n", pod.Name, pod.Namespace)
		return nil
	}

	for i := 0; i < MAX_RETRIES; i++ {
		deployment, err := clientset.AppsV1().Deployments(pod.Namespace).Get(context.TODO(), deploymentName, metav1.GetOptions{})
		if err != nil {
			log.Printf("Error getting Deployment %s in %s : %v\n", deploymentName, pod.Namespace, err)
			return err
		}

		replicas := int32(0)
		deployment.Spec.Replicas = &replicas

		_, err = clientset.AppsV1().Deployments(pod.Namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{})
		if err != nil {
			log.Printf("Attempt %d: Error updating replicas to 0 for %s in %s : %v\n", i+1, deploymentName, pod.Namespace, err)
			time.Sleep(5 * time.Second)
		} else {
			log.Printf("Successfully scaled down %s in %s \n", deploymentName, pod.Namespace)
			return nil
		}
	}

	return fmt.Errorf("failed to scale down %s in %s", deploymentName, pod.Namespace)
}
