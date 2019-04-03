package cmd

import (
	"github.com/kubextender/pvcexec/pkg/k8s"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
)

var (
	zshImageName = "kubextender/pvcexec-zsh:LATEST"
	zshPodName   = "pvcexec-zsh"
)

type ZshOptions struct {
	pvcExecOptions *k8s.PvcExecOptions
}

func NewZshOptions(options *k8s.PvcExecOptions) *ZshOptions {
	return &ZshOptions{
		pvcExecOptions: options,
	}
}
func NewZshCommand(pvcexecOptions *k8s.PvcExecOptions) *cobra.Command {
	options := NewZshOptions(pvcexecOptions)
	cmd := &cobra.Command{
		Use:                   "zsh",
		Short:                 "Mounts provided pvc(s) to the new pod and run zsh shell",
		Example:               "kubectl pvcexec zsh -pvc testpvc1 -pvc testpvc2 -pvc testpvc3",
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
	cmd.Flags().StringP("namespace", "n", "", "use this flag to override kubernetes namespace from current kubectl context")
	return cmd
}

// Complete completes the setup of the command.
func (zshOptions *ZshOptions) complete(cmd *cobra.Command, args []string) error {
	var err error
	options := zshOptions.pvcExecOptions
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
	options.Command = []string{"/bin/zsh"}
	options.ImageName = zshImageName
	options.PodName = zshPodName
	return nil
}

func (zshOptions *ZshOptions) run() error {
	options := zshOptions.pvcExecOptions
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
