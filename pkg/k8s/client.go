package k8s

import (
	"k8s.io/client-go/kubernetes"
	"log"

	// Ensure we have auth plugins (gcp, azure, openstack, ...) linked in
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func New(url string, ca []byte, token string) *kubernetes.Clientset {
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
	// TODO() Move qps and burst in config
	restConfig.QPS = 60
	restConfig.Burst = 120

	clientSet, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		log.Fatalf("failed to build akubernetes  NewForConfig: %v", err)
		return nil
	}

	return clientSet
}
