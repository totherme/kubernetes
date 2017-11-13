package cmd

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"k8s.io/kubernetes/pkg/kubectl/cmd/templates"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/kubectl/resource"
	"k8s.io/kubernetes/pkg/kubectl/util/i18n"
)

var (
	helloWorldMsg  string = "Hello World\n"
	helloWorldVerb string = "hello-world"

	helloWorldLong = templates.LongDesc(i18n.T(`
		Print a friendly message on the console, to welcome the user into the
		wonderful world of kubernetes.

		This exercise comes from:
		https://github.com/kubernetes/community/blob/master/sig-cli/CONTRIBUTING.md`))

	helloKubernetesVerb string = "hello-kubernetes"
	helloKubernetesLong string = templates.LongDesc(i18n.T(`
		A friendly way of finding out the kind and name of some resource contained in a file.

		Say which file(s) you want to inspect by using the -f flags. This will print their kinds and names.
	`))
)

func NewCmdHelloWorld(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   helloWorldVerb,
		Short: i18n.T("says hello to the world"),
		Long:  helloWorldLong,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprint(out, helloWorldMsg)
		},
	}
	return cmd
}

func NewCmdHelloKubernetes(out io.Writer) *cobra.Command {
	var options resource.FilenameOptions

	cmd := &cobra.Command{
		Use:   helloKubernetesVerb,
		Short: i18n.T("says hello to some k8s resourse"),
		Long:  helloKubernetesLong,
		Run: func(cmd *cobra.Command, args []string) {
		},
	}

	cmdutil.AddFilenameOptionFlags(cmd, &options, "do awesome things")

	return cmd
}
