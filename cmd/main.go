package main

import (
	"flag"
	"golang.org/x/crypto/ssh/terminal"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/client-go/util/homedir"
	"os"
	"path/filepath"
	"strings"
)

const podName = "mcpvc"
const imageName = "kodiraj.ga:5001/mcpvc/runner-docker:latest"

func main() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "absolute path to the kubeconfig file - (optional)")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file - (required)")
	}
	var namespace = flag.String("n", "default", "namespace - (optional)")
	var pvcsArg = flag.String("pvcs", "", "list of pvc names to mount under / folder, i.e. -pvcs 'pvc1 pvc2 pvc3' - (at least one required)")
	flag.Parse()
	if *pvcsArg == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}
	var pvcNames = strings.Split(*pvcsArg, " ")
	secondDir := ""
	if len(pvcNames) > 1 {
		secondDir = pvcNames[1]
	}
	var command = []string{"/usr/bin/mc", pvcNames[0], secondDir}
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	pod := createRunnerPod(clientset, *namespace, pvcNames)
	// delete pod on exit
	defer clientset.CoreV1().Pods(*namespace).Delete(podName, metav1.NewDeleteOptions(0))
	execCommandInPod(clientset, pod, command, err, config)
}

func execCommandInPod(clientset *kubernetes.Clientset, pod *apiv1.Pod, command []string, err error, config *rest.Config) {
	req := clientset.CoreV1().RESTClient().
		Post().
		Namespace(pod.Namespace).
		Resource("pods").
		Name(pod.Name).
		SubResource("exec").
		VersionedParams(&apiv1.PodExecOptions{
			Container: pod.Spec.Containers[0].Name,
			Command:   command,
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
			TTY:       true,
		}, scheme.ParameterCodec)
	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
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

func createRunnerPod(clientset *kubernetes.Clientset, namespace string, pvcs []string) *apiv1.Pod {
	pods := clientset.CoreV1().Pods(namespace)
	volumes := make([]apiv1.Volume, len(pvcs))
	volumeMounts := make([]apiv1.VolumeMount, len(pvcs))
	for i, pvcName := range pvcs {
		volumeMounts[i] = apiv1.VolumeMount{MountPath: "/" + pvcName, Name: pvcName}
		volumes[i] = apiv1.Volume{Name: pvcName, VolumeSource: apiv1.VolumeSource{PersistentVolumeClaim: &apiv1.PersistentVolumeClaimVolumeSource{ClaimName: pvcName, ReadOnly: false}}}
	}
	mcpvcPod := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: namespace,
		},
		Spec: apiv1.PodSpec{
			Volumes: volumes,
			Containers: []apiv1.Container{
				{
					Name:    podName,
					Image:   imageName,
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
	pod, err := pods.Create(mcpvcPod)
	if err != nil {
		panic(err)
	}
	// Wait for the Pod to indicate Ready == True.
	watcher, err := pods.Watch(
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
