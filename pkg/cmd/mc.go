package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes/scheme"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/tools/remotecommand"

	"os"
)

type RunOptions struct {
	configFlags *genericclioptions.ConfigFlags
	namespace   string
	rawConfig   api.Config
	genericclioptions.IOStreams
	restConfig *rest.Config
	podName    string
	pvcNames   []string
	imageName  string
	command    []string
}

func NewRunOptions(streams genericclioptions.IOStreams) *RunOptions {
	return &RunOptions{
		configFlags: genericclioptions.NewConfigFlags(true),
		IOStreams:   streams,
	}
}

func NewRootCmd(streams genericclioptions.IOStreams) *cobra.Command {
	opt := NewRunOptions(streams)
	cmd := &cobra.Command{
		Use:          "mc -pvc pvcname1 -pvc pvcname2]",
		Short:        "Mounts provided pvc to pod and run Midnight Commander",
		Example:      fmt.Sprintf("", "kubectl"),
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := opt.Complete(c, args); err != nil {
				return err
			}
			if err := opt.Run(); err != nil {
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
func (opt *RunOptions) Complete(cmd *cobra.Command, args []string) error {
	// Prepare namespace
	var err error
	opt.namespace, _, err = opt.configFlags.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		return err
	}
	// Prepare client
	opt.restConfig, err = opt.configFlags.ToRESTConfig()
	if err != nil {
		return err
	}
	opt.pvcNames, _ = cmd.Flags().GetStringArray("pvc")
	if err != nil {
		return err
	}

	secondDir := ""
	if len(opt.pvcNames) > 1 {
		secondDir = opt.pvcNames[1]
	}
	opt.command = []string{"/usr/bin/mc", opt.pvcNames[0], secondDir}
	opt.imageName = "kodiraj.ga:5001/mcpvc/runner-docker:latest"
	opt.podName = "mcpvc"
	return nil
}

func (opt *RunOptions) Run() error {
	restConfig, _ := opt.configFlags.ToRESTConfig()
	podClient, _ := corev1client.NewForConfig(restConfig)
	pod := createRunnerPod(podClient, opt)
	// delete pod on exit
	defer podClient.Pods(opt.namespace).Delete(opt.podName, metav1.NewDeleteOptions(0))
	execCommandInPod(podClient, pod, opt)
	return nil
}

func execCommandInPod(coreClient *corev1client.CoreV1Client, pod *apiv1.Pod, opt *RunOptions) {
	req := coreClient.RESTClient().
		Post().
		Namespace(pod.Namespace).
		Resource("pods").
		Name(pod.Name).
		SubResource("exec").
		VersionedParams(&apiv1.PodExecOptions{
			Container: pod.Spec.Containers[0].Name,
			Command:   opt.command,
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
			TTY:       true,
		}, scheme.ParameterCodec)
	exec, err := remotecommand.NewSPDYExecutor(opt.restConfig, "POST", req.URL())
	if err != nil {
		panic(err.Error())
	}
	// Put the terminal into raw mode to prevent it echoing characters twice.
	oldState, err := terminal.MakeRaw(0)
	if err != nil {
		panic(err.Error())
	}
	defer terminal.Restore(0, oldState)
	// Connect this process' std{in,out,err} to the remote shell process.
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Tty:    true,
	})
	if err != nil {
		panic(err.Error())
	}
}

func createRunnerPod(pods *corev1client.CoreV1Client, options *RunOptions) *apiv1.Pod {
	volumes := make([]apiv1.Volume, len(options.pvcNames))
	volumeMounts := make([]apiv1.VolumeMount, len(options.pvcNames))
	for i, pvcName := range options.pvcNames {
		volumeMounts[i] = apiv1.VolumeMount{MountPath: "/" + pvcName, Name: pvcName}
		volumes[i] = apiv1.Volume{Name: pvcName, VolumeSource: apiv1.VolumeSource{PersistentVolumeClaim: &apiv1.PersistentVolumeClaimVolumeSource{ClaimName: pvcName, ReadOnly: false}}}
	}
	mcpvcPod := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      options.podName,
			Namespace: options.namespace,
		},
		Spec: apiv1.PodSpec{
			Volumes: volumes,
			Containers: []apiv1.Container{
				{
					Name:    options.podName,
					Image:   options.imageName,
					Command: []string{"cat"},
					Stdin:   true,
					Env: []apiv1.EnvVar{
						{Name: "TERM", Value: getEnv("TERM", "xterm-256color")},
						{Name: "COLUMNS", Value: getEnv("COLUMNS", "180")},
						{Name: "LINES", Value: getEnv("LINES", "60")},
					},
					VolumeMounts: volumeMounts,
				},
			},
			ImagePullSecrets: []apiv1.LocalObjectReference{
				{Name: "regcred"},
			}},
	}
	podsClient := pods.Pods(options.namespace)
	pod, err := podsClient.Create(mcpvcPod)
	if err != nil {
		panic(err)
	}
	// Wait for the Pod to indicate Ready == True.
	watcher, err := podsClient.Watch(
		metav1.SingleObject(pod.ObjectMeta),
	)
	if err != nil {
		panic(err)
	}
	for event := range watcher.ResultChan() {
		switch event.Type {
		case watch.Modified:
			pod = event.Object.(*apiv1.Pod)
			// If the Pod contains a status condition Ready == True, stop
			// watching.
			for _, cond := range pod.Status.Conditions {
				if cond.Type == apiv1.PodReady &&
					cond.Status == apiv1.ConditionTrue {
					watcher.Stop()
				}
			}
		default:
			panic("unexpected event type " + event.Type)
		}
	}
	return pod
}
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return defaultValue
	}
	return value
}
