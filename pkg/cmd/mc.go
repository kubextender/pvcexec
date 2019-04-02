package cmd

import (
	"github.com/kubextender/pvcexec/pkg/k8s"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
)

var (
	ImageName = "kubextender/pvcexec-mc:latest"
	PodName   = "pvcexec-mc"
)

type McOptions struct {
	pvcExecOptions *k8s.PvcExecOptions
}

func NewMcOptions(options *k8s.PvcExecOptions) *McOptions {
	return &McOptions{
		pvcExecOptions: options,
	}
}
func NewMcCommand(pvcexecOptions *k8s.PvcExecOptions) *cobra.Command {
	options := NewMcOptions(pvcexecOptions)
	cmd := &cobra.Command{
		Use:                   "mc (list of pvcs to mount)",
		Short:                 "Mounts provided pvc(s) to the new pod and run Midnight Commander",
		Example:               "kubectl pvcexec mc -pvc testpvc1 -pvc testpvc2 -pvc testpvc3",
		DisableFlagsInUseLine: true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := options.Complete(c, args); err != nil {
				return err
			}
			if err := options.Run(); err != nil {
				return err
			}
			return nil
		},
	}
	cmd.Flags().StringArrayP("pvc", "p", nil, "-pvc pvc1 -pvc pvc2 ...")
	cmd.MarkFlagRequired("pvc")
	cmd.Flags().StringP("namespace", "n", "", "use this flag to override kubernetes namespace from current kubectl context")
	return cmd
}

// Complete completes the setup of the command.
func (mcOptions *McOptions) Complete(cmd *cobra.Command, args []string) error {
	var err error
	options := mcOptions.pvcExecOptions
	// Prepare namespace
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
	options.PvcNames, _ = cmd.Flags().GetStringArray("pvc")
	if err != nil {
		return err
	}
	secondDir := ""
	if len(options.PvcNames) > 1 {
		secondDir = options.PvcNames[1]
	}
	options.Command = []string{"/usr/bin/mc", options.PvcNames[0], secondDir}
	options.ImageName = ImageName
	options.PodName = PodName
	return nil
}

func (mcOptions *McOptions) Run() error {
	options := mcOptions.pvcExecOptions
	restConfig, _ := options.ConfigFlags.ToRESTConfig()
	podClient, _ := corev1client.NewForConfig(restConfig)

	defer podClient.Pods(options.Namespace).Delete(options.PodName, metav1.NewDeleteOptions(0))
	pod, err := k8s.CreateRunnerPod(podClient, options)
	if err != nil {
		return err
	}
	// delete pod on exit
	k8s.ExecCommandInPod(podClient, pod, options)
	return nil
}
