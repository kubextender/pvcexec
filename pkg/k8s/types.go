package k8s

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd/api"
	"time"
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
	Timeout    time.Duration
}

func (options *PvcExecOptions) Complete(cmd *cobra.Command, args []string) error {
	var err error
	options.Namespace, _, err = options.ConfigFlags.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		return err
	}

	var overrideNamespace, _ = cmd.Flags().GetString("namespace")
	if len(overrideNamespace) > 0 {
		options.Namespace = overrideNamespace
	}
	// Prepare client
	options.RestConfig, err = options.ConfigFlags.ToRESTConfig()
	if err != nil {
		return err
	}
	options.Timeout, err = cmd.Flags().GetDuration("timeout")
	if err != nil {
		return err
	}

	return nil
}
