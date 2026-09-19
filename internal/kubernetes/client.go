package kubernetes

import (
	"fmt"
	"os"

	k8s "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Options control how KubeCone authenticates to a cluster.
type Options struct {
	Kubeconfig string
	Context    string
	Namespace  string
}

// Client is the Kubernetes access layer. Investigation code talks to this,
// not to client-go types scattered across the repo.
type Client struct {
	Kube      k8s.Interface
	REST      *rest.Config
	Namespace string
	Context   string
}

// New builds a clientset from kubeconfig, matching kubectl's resolution order.
func New(opts Options) (*Client, error) {
	loading := clientcmd.NewDefaultClientConfigLoadingRules()
	if opts.Kubeconfig != "" {
		loading.ExplicitPath = opts.Kubeconfig
	} else if env := os.Getenv("KUBECONFIG"); env != "" {
		loading.ExplicitPath = env
	}

	overrides := &clientcmd.ConfigOverrides{}
	if opts.Context != "" {
		overrides.CurrentContext = opts.Context
	}

	cfg := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loading, overrides)
	restCfg, err := cfg.ClientConfig()
	if err != nil {
		if restCfg, err = rest.InClusterConfig(); err != nil {
			return nil, fmt.Errorf("load kubeconfig: %w", err)
		}
	}

	clientset, err := k8s.NewForConfig(restCfg)
	if err != nil {
		return nil, fmt.Errorf("create clientset: %w", err)
	}

	ns := opts.Namespace
	if ns == "" {
		ns, _, err = cfg.Namespace()
		if err != nil || ns == "" {
			ns = "default"
		}
	}

	ctxName := opts.Context
	if ctxName == "" {
		raw, err := cfg.RawConfig()
		if err == nil {
			ctxName = raw.CurrentContext
		}
	}

	return &Client{
		Kube:      clientset,
		REST:      restCfg,
		Namespace: ns,
		Context:   ctxName,
	}, nil
}

// NewForInterface is used by tests that inject a fake clientset.
func NewForInterface(kube k8s.Interface, namespace string) *Client {
	if namespace == "" {
		namespace = "default"
	}
	return &Client{Kube: kube, Namespace: namespace, Context: "fake"}
}
