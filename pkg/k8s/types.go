package k8s

import (
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd/api"
)

type PvcExecOptions struct {
	*genericclioptions.ConfigFlags
	Namespace string
	RawConfig api.Config
	genericclioptions.IOStreams
	RestConfig *rest.Config
	PvcNames   []string
	PodName    string
	ImageName  string
	Command    []string
}
