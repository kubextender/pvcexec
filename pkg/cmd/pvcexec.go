package cmd

import (
	"fmt"
	"github.com/kubextender/pvcexec/pkg/k8s"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func NewPvcExecOptions(streams genericclioptions.IOStreams) *k8s.PvcExecOptions {
	flags := genericclioptions.NewConfigFlags(true)
	var options = &k8s.PvcExecOptions{
		ConfigFlags: flags,
		IOStreams:   streams,
	}
	return options
}

func NewPvcExecCmd(streams genericclioptions.IOStreams) *cobra.Command {
	o := NewPvcExecOptions(streams)
	cmd := &cobra.Command{
		Use:                   "pvcexec [flags] [command]",
		Short:                 "Mounts provided pvc(s) to the new pod and run command",
		Example:               "pvcexec -n default mc",
		DisableFlagsInUseLine: true,
		PersistentPreRun: func(c *cobra.Command, args []string) {
			if err := o.Complete(c, args); err != nil {
				fmt.Errorf("ERROR: can't determine namespace\n")
				c.SetOutput(streams.ErrOut)
				cobra.NoArgs(c, args)
				c.Help()
			}
			fmt.Printf("Selected namespace: %s\n", o.Namespace)
		},
	}
	cmd.PersistentFlags().StringP("namespace", "n", "", "If present, the namespace scope for this CLI request.")
	cmd.AddCommand(NewMcCommand(o))
	cmd.AddCommand(NewZshCommand(o))
	cmd.AddCommand(NewVersionCommand(streams))
	cmd.AddCommand(NewListCommand(o))
	return cmd
}
