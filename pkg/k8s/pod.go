package k8s

import (
	"fmt"
	"golang.org/x/crypto/ssh/terminal"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/scheme"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/remotecommand"
	"os"
)

func ExecCommandInPod(coreClient *corev1client.CoreV1Client, pod *apiv1.Pod, opt *PvcExecOptions) {
	req := coreClient.RESTClient().
		Post().
		Namespace(pod.Namespace).
		Resource("pods").
		Name(pod.Name).
		SubResource("exec").
		VersionedParams(&apiv1.PodExecOptions{
			Container: pod.Spec.Containers[0].Name,
			Command:   opt.Command,
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
			TTY:       true,
		}, scheme.ParameterCodec)
	exec, err := remotecommand.NewSPDYExecutor(opt.RestConfig, "POST", req.URL())
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

func CreateRunnerPod(pods *corev1client.CoreV1Client, options *PvcExecOptions) (*apiv1.Pod, error) {
	volumes := make([]apiv1.Volume, len(options.PvcNames))
	volumeMounts := make([]apiv1.VolumeMount, len(options.PvcNames))
	for i, pvcName := range options.PvcNames {
		volumeMounts[i] = apiv1.VolumeMount{MountPath: "/" + pvcName, Name: pvcName}
		claimVolumeSource := &apiv1.PersistentVolumeClaimVolumeSource{ClaimName: pvcName, ReadOnly: false}
		volumes[i] = apiv1.Volume{Name: pvcName, VolumeSource: apiv1.VolumeSource{PersistentVolumeClaim: claimVolumeSource}}
	}
	mcpvcPod := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      options.PodName,
			Namespace: options.Namespace,
		},
		Spec: apiv1.PodSpec{
			Volumes: volumes,
			Containers: []apiv1.Container{
				{
					Name:    options.PodName,
					Image:   options.ImageName,
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
	podsClient := pods.Pods(options.Namespace)
	pod, err := podsClient.Create(mcpvcPod)
	if err != nil {
		return nil, err
	}
	// Wait for the Pod to indicate Ready == True.
	watcher, err := podsClient.Watch(
		metav1.SingleObject(pod.ObjectMeta),
	)
	if err != nil {
		return nil, err
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
				} else if cond.Status == apiv1.ConditionFalse && cond.Reason == apiv1.PodReasonUnschedulable {
					watcher.Stop()
					return nil, fmt.Errorf("Error %v", cond.Message)
				}
			}
		default:
			panic("unexpected event type " + event.Type)
		}
	}
	return pod, nil
}
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return defaultValue
	}
	return value
}
