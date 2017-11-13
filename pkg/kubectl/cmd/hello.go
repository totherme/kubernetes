package cmd

import (
	"fmt"
	"io"
	"github.com/spf13/cobra"
	"k8s.io/kubernetes/pkg/kubectl/util/i18n"
	"k8s.io/kubernetes/pkg/kubectl/cmd/templates"
)

var (
	helloWorldMsg string = "Hello World\n"
	helloWorldVerb string = "hello-world"

	helloWorldLong = templates.LongDesc(i18n.T(`
		Print a friendly message on the console, to welcome the user into the
		wonderful world of kubernetes.

		This exercise comes from:
		https://github.com/kubernetes/community/blob/master/sig-cli/CONTRIBUTING.md`))

)

func NewCmdHelloWorld(out io.Writer) *cobra.Command {
	 cmd := &cobra.Command{
		Use: helloWorldVerb,
		Short: i18n.T("says hello to the world"),
		Long: helloWorldLong,
		Run: func(cmd *cobra.Command, args []string) {
			sayHello(out)
		},
	 }
	 return cmd
}

func sayHello(out io.Writer) {
	fmt.Fprint(out, helloWorldMsg)
}