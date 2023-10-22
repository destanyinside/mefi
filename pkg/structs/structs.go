package structs

import (
	"k8s.io/client-go/kubernetes"
)

type Clusters struct {
	Clusters []Configurations `yaml:"clusters"`
}

type Type string

const (
	Remote Type = "remote"
	Local  Type = "local"
)

type Configurations struct {
	Url   string `yaml:"url"`
	Token string `yaml:"token"`
	Ca    string `yaml:"ca"`
	Name  string `yaml:"name"`
	Type  Type   `yaml:"type"`
}

type K8sClients struct {
	K8sCli []K8sClient
}

type K8sClient struct {
	ClusterName string
	ClientSet   *kubernetes.Clientset
}
