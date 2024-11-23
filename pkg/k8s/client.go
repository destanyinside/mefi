package k8s

import (
	"github.com/destanyinside/mefi/pkg/log"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func New(url string, ca []byte, token string, log *log.LogrusLogger, qps float32, burst int) *kubernetes.Clientset {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	config := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules,
		&clientcmd.ConfigOverrides{
			AuthInfo: api.AuthInfo{
				Token: token,
			},
			ClusterInfo: clientcmdapi.Cluster{
				Server:                   url,
				CertificateAuthorityData: ca,
			},
		})

	restConfig, err := config.ClientConfig()
	if err != nil {
		log.Fatalf("failed to build a kubernetes ClientConfig: %v", err)
		return nil
	}
	restConfig.QPS = qps
	restConfig.Burst = burst

	clientSet, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		log.Fatalf("failed to build akubernetes  NewForConfig: %v", err)
		return nil
	}

	return clientSet
}
