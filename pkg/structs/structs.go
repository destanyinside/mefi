package structs

import "k8s.io/client-go/dynamic"

type Clusters struct {
	Clusters []Configurations `yaml:"clusters"`
}

type Configurations struct {
	Server string `yaml:"server"`
	Token  string `yaml:"token"`
	Ca     string `yaml:"CA"`
	Name   string `yaml:"name"`
}

type K8sClients struct {
	K8sCli []K8sClient
}

type K8sClient struct {
	ClusterName string
	DInf        dynamic.Interface
}
