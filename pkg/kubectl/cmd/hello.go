package cmd

import (
	"fmt"
	"io"

	"io/ioutil"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/kubernetes/pkg/kubectl/cmd/templates"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/kubectl/resource"
	"k8s.io/kubernetes/pkg/kubectl/util/i18n"
	"k8s.io/kubernetes/staging/src/k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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
		RunE: func(cmd *cobra.Command, args []string) error {
			file, err := os.Open(options.Filenames[0])
			defer file.Close()
			if err != nil {
				return fmt.Errorf("could not open file in HelloKubernetes: %s", err)
			}
			bytes, err := ioutil.ReadAll(file)
			if err != nil {
				return fmt.Errorf("could not read file in HelloKubernets: %s", err)
			}
			object, kind, err := unstructured.UnstructuredJSONScheme.Decode(bytes, nil, nil)
			if err != nil {
				return fmt.Errorf("could not decode file in HelloKubernetes: %s", err)
			}
			unstructuredObject, ok := object.(*unstructured.Unstructured)
			if !ok {
				return fmt.Errorf("decoded file in HelloKubernetes was the wrong type for me to get its name")
			}

			fmt.Fprintf(out, "Hello %s %s", kind.Kind, unstructuredObject.GetName())
			return nil
		},
	}

	cmdutil.AddFilenameOptionFlags(cmd, &options, "do awesome things")

	return cmd
}
