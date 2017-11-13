package cmd

import (
	"fmt"
	"io"
	"github.com/spf13/cobra"
)

var (
	helloWorldMsg string = "Hello World\n"
	helloWorldVerb string = "hello-world"
)

func NewCmdHelloWorld(out io.Writer) *cobra.Command {
	 cmd := &cobra.Command{
		Use: helloWorldVerb,
		Run: func(cmd *cobra.Command, args []string) {
			sayHello(out)
		},
	 }
	 return cmd
}

func sayHello(out io.Writer) {
	fmt.Fprint(out, helloWorldMsg)
}