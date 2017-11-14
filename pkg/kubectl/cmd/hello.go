package cmd

import (
	"fmt"
	"io"

	"os"

	"github.com/spf13/cobra"
	"k8s.io/kubernetes/pkg/kubectl/cmd/templates"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/kubectl/resource"
	"k8s.io/kubernetes/pkg/kubectl/util/i18n"
	"k8s.io/kubernetes/staging/src/k8s.io/apimachinery/pkg/util/yaml"
)

var (
	helloWorldMsg   string = "Hello World\n"
	helloWorldUsage string = "hello-world"

	helloWorldLong = templates.LongDesc(i18n.T(`
		Print a friendly message on the console, to welcome the user into the
		wonderful world of kubernetes.

		This exercise comes from:
		https://github.com/kubernetes/community/blob/master/sig-cli/CONTRIBUTING.md`))

	helloKubernetesUsage string = "hello-kubernetes -f FILENAME"
	helloKubernetesLong  string = templates.LongDesc(i18n.T(`
		A friendly way of finding out the kind and name of some resource contained in a file.

		Say which file(s) you want to inspect by using the -f flags. This will print their kinds and names.
	`))
)

func NewCmdHelloWorld(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   helloWorldUsage,
		Short: i18n.T("says hello to the world"),
		Long:  helloWorldLong,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprint(out, helloWorldMsg)
		},
	}
	return cmd
}

type FailureHandler func(c *cobra.Command, args []string)

func NewCmdHelloKubernetes(out, errOut io.Writer, handler FailureHandler) *cobra.Command {
	var options resource.FilenameOptions

	cmd := &cobra.Command{
		Use:   helloKubernetesUsage,
		Short: i18n.T("says hello to some k8s resourse"),
		Long:  helloKubernetesLong,
		Run: func(cmd *cobra.Command, args []string) {
			if cmdutil.IsFilenameSliceEmpty(options.Filenames) {
				handler(cmd, args)
				return
			}

			for _, filename := range options.Filenames {
				file, err := os.Open(filename)
				defer file.Close()
				cmdutil.CheckErr(err)

				decoder := yaml.NewYAMLOrJSONDecoder(file, 256)
				var object map[string]interface{}
				err = decoder.Decode(&object)
				cmdutil.CheckErr(err)

				name := object["metadata"].(map[string]interface{})["name"]
				kind := object["kind"]

				fmt.Fprintf(out, "Hello %s %s\n", kind, name)
			}
		},
	}

	cmdutil.AddFilenameOptionFlags(cmd, &options, "do awesome things")
	cmd.MarkFlagRequired("filename")

	return cmd
}
