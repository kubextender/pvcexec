package k8s

import (
	"fmt"
	"golang.org/x/crypto/ssh/terminal"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/remotecommand"
	"time"

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
		volumeMounts[i] = apiv1.VolumeMount{MountPath: "/mnt/" + pvcName, Name: pvcName}
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
					ImagePullPolicy: apiv1.PullAlways,
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
	status := pod.Status
	w, err := podsClient.Watch(
		metav1.SingleObject(pod.ObjectMeta),
	)
	if err != nil {
		return nil, err
	}
	func() {
		for {
			select {
			case events, ok := <-w.ResultChan():
				if !ok {
					return
				}
				pod = events.Object.(*apiv1.Pod)
				fmt.Println("Checking pod status:", status.Phase)
				status = pod.Status
				if pod.Status.Phase != apiv1.PodPending {
					w.Stop()
				}
			case <-time.After(15 * time.Second):
				fmt.Println("timeout to wait for pod active")
				w.Stop()
			}
		}
	}()
	if status.Phase != apiv1.PodRunning {
		return nil, fmt.Errorf("Pod is unavailable: %v", status.Phase)
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
