package cmd

import (
	"github.com/kubextender/pvcexec/pkg/k8s"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func NewPvcExecOptions(streams genericclioptions.IOStreams) *k8s.PvcExecOptions {
	options := &k8s.PvcExecOptions{
		ConfigFlags: genericclioptions.NewConfigFlags(true),
		IOStreams:   streams,
	}
	return options
}

func NewPvcExecCmd(streams genericclioptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "pvcexec COMMAND=mc|ohmyzsh|bash [options]",
		Short:                 "Mounts provided pvc(s) to the new pod and run command",
		Example:               "pvcexec [sub-command]",
		DisableFlagsInUseLine: true,
		Run: func(c *cobra.Command, args []string) {
			c.SetOutput(streams.ErrOut)
			cobra.NoArgs(c, args)
			c.Help()
		},
	}
	o := NewPvcExecOptions(streams)
	cmd.AddCommand(NewMcCommand(o))
	cmd.AddCommand(NewZshCommand(o))
	return cmd
}
