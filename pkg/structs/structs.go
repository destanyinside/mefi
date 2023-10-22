package structs

import "k8s.io/client-go/dynamic"

type Clusters struct {
	Clusters []Configurations `yaml:"clusters"`
}

type Configurations struct {
	Url   string `yaml:"url"`
	Token string `yaml:"token"`
	Ca    string `yaml:"ca"`
	Name  string `yaml:"name"`
	Type  string `yaml:"type"`
}

type K8sClients struct {
	K8sCli []K8sClient
}

type K8sClient struct {
	ClusterName string
	DInf        dynamic.Interface
}
