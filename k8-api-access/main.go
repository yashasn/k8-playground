package main

import (
	"context"
	"fmt"
	"log"

	"github.com/yashasn/k8s-api-access/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func main() {
	clientset, err := client.NewK8sClient()
	if err != nil {
		log.Fatalf("Failed to create K8s client: %v", err)
	}

	// Use the clientset to interact with the Kubernetes API
	pods, err := clientset.CoreV1().Pods("default").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Failed to list pods: %v", err)
	}

	for _, pod := range pods.Items {
		fmt.Printf("Pod Name: %s\n", pod.Name)
	}
}
