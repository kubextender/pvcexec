package cmd

import (
	"fmt"
	"github.com/kubextender/pvcexec/pkg/k8s"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"os"
	"text/tabwriter"
)

func NewListCommand(options *k8s.PvcExecOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List existing pvcs",
		RunE: func(c *cobra.Command, args []string) error {
			restConfig, _ := options.ConfigFlags.ToRESTConfig()
			client, _ := corev1client.NewForConfig(restConfig)
			pvcList, e := client.PersistentVolumeClaims(options.Namespace).List(metav1.ListOptions{})
			if e != nil {
				panic(e)
			}
			// initialize tabwriter
			w := new(tabwriter.Writer)
			// minwidth, tabwidth, padding, padchar, flags
			w.Init(os.Stdout, 8, 8, 0, '\t', 0)
			defer w.Flush()
			_, _ = fmt.Fprintf(w, "\n %s\t%s\t", "Name", "Phase")
			_, _ = fmt.Fprintf(w, "\n %s\t%s\t", "------", "------")

			for _, it := range pvcList.Items {
				_, _ = fmt.Fprintf(w, "\n %s\t%s", it.Name, it.Status.Phase)
			}
			return nil
		},
	}
	return cmd
}
