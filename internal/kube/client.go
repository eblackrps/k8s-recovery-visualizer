package kube

import (
	"fmt"
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func NewClient(kubeconfigPath string) (*kubernetes.Clientset, *rest.Config, error) {
	var (
		cfg *rest.Config
		err error
	)

	if kubeconfigPath != "" {
		cfg, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	} else {
		cfg, err = rest.InClusterConfig()
		if err != nil {
			if home, ok := os.LookupEnv("HOME"); ok {
				cfg, err = clientcmd.BuildConfigFromFlags("", home+"/.kube/config")
			}
		}
	}

	if err != nil {
		return nil, nil, fmt.Errorf("load kube config: %w", err)
	}

	cs, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("new clientset: %w", err)
	}

	return cs, cfg, nil
}
