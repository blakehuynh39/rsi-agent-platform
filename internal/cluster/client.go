package cluster

import (
	"fmt"
	"strings"

	"github.com/piplabs/rsi-agent-platform/internal/config"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func NewConfig(cfg config.Config) (*rest.Config, error) {
	if strings.TrimSpace(cfg.KubeconfigPath) != "" {
		loadingRules := &clientcmd.ClientConfigLoadingRules{ExplicitPath: cfg.KubeconfigPath}
		overrides := &clientcmd.ConfigOverrides{}
		if strings.TrimSpace(cfg.KubernetesContext) != "" {
			overrides.CurrentContext = cfg.KubernetesContext
		}
		clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, overrides)
		restConfig, err := clientConfig.ClientConfig()
		if err != nil {
			return nil, fmt.Errorf("load kubeconfig: %w", err)
		}
		return restConfig, nil
	}
	restConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("load in-cluster config: %w", err)
	}
	return restConfig, nil
}

func NewClientset(cfg config.Config) (*kubernetes.Clientset, error) {
	restConfig, err := NewConfig(cfg)
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("create kubernetes client: %w", err)
	}
	return clientset, nil
}
