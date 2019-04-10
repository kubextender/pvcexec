package cmd

import (
	"github.com/kubextender/pvcexec/pkg/k8s"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
)

var (
	mcImageName = "kubextender/pvcexec-mc:LATEST"
	mcPodName   = "pvcexec-mc"
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
		Use:                   "mc",
		Short:                 "Mounts provided pvc(s) to the new pod and run Midnight Commander",
		Example:               "kubectl pvcexec mc -pvc testpvc1 -pvc testpvc2 -pvc testpvc3",
		DisableFlagsInUseLine: true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := options.complete(c, args); err != nil {
				return err
			}
			if err := options.run(); err != nil {
				return err
			}
			return nil
		},
	}
	cmd.Flags().StringArrayP("pvc", "p", nil, "-pvc pvc1 -pvc pvc2 ...")
	cmd.MarkFlagRequired("pvc")
	return cmd
}

// Complete completes the setup of the command.
func (mcOptions *McOptions) complete(cmd *cobra.Command, args []string) error {
	options := mcOptions.pvcExecOptions
	options.PvcNames, _ = cmd.Flags().GetStringArray("pvc")
	secondDir := "/mnt/"
	if len(options.PvcNames) > 1 {
		secondDir += options.PvcNames[1]
	}
	options.Command = []string{"/usr/bin/mc", "-S", "gotar", "/mnt/" + options.PvcNames[0], secondDir}
	options.ImageName = mcImageName
	options.PodName = mcPodName
	return nil
}

func (mcOptions *McOptions) run() error {
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
