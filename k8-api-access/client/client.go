package client

import (
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// NewK8sClient creates a new Kubernetes client.
func NewK8sClient() (*kubernetes.Clientset, error) {
	kubeconfig := getKubeConfigPath()

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

// getKubeConfigPath returns the path to the kubeconfig file.
func getKubeConfigPath() string {
	// If we're running in a local environment, use the default kubeconfig location
	if home := homedir.HomeDir(); home != "" {
		fmt.Println(filepath.Join(home, ".kube", "config"))
		return filepath.Join(home, ".kube", "config")
	}
	// Fallback to the environment variable or default path
	return os.Getenv("KUBECONFIG")
}
