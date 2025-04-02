package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	kubeconfig := flag.String("kubeconfig", "", "Path to a kubeconfig. Only required if running outside the cluster.")
	flag.Parse()

	var config *rest.Config
	var err error
	if *kubeconfig == "" {
		config, err = rest.InClusterConfig()
	} else {
		config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
	}
	if err != nil {
		panic(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	// Create shared informer factory
	factory := informers.NewSharedInformerFactory(clientset, 30*time.Second)
	podInformer := factory.Core().V1().Pods().Informer()

	stopCh := make(chan struct{})
	defer close(stopCh)

	podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pod, ok := obj.(*v1.Pod)
			if !ok {
				return
			}

			// Check if pod has the annotation
			if value, exists := pod.Annotations["auto-label"]; exists && value == "true" {
				addLabelToPod(clientset, pod)
			}
		},
	})

	factory.Start(stopCh)
	wait.Until(func() { factory.WaitForCacheSync(stopCh) }, time.Second, stopCh)
}

// addLabelToPod updates the Pod with a new label
func addLabelToPod(clientset *kubernetes.Clientset, pod *v1.Pod) {
	// Add the label
	patch := `{"metadata": {"labels": {"managed-by": "custom-controller"}}}`
	_, err := clientset.CoreV1().Pods(pod.Namespace).Patch(context.TODO(), pod.Name, types.MergePatchType, []byte(patch), metav1.PatchOptions{})
	if err != nil {
		fmt.Printf("Failed to label pod %s: %v\n", pod.Name, err)
		return
	}
	fmt.Printf("Labeled pod %s in namespace %s\n", pod.Name, pod.Namespace)
}
