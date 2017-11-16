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

func NewCmdHelloKubernetes(f cmdutil.Factory, out, errOut io.Writer, handler FailureHandler) *cobra.Command {
	var filesOpts resource.FilenameOptions

	cmd := &cobra.Command{
		Use:   helloKubernetesUsage,
		Short: i18n.T("says hello to some k8s resourse"),
		Long:  helloKubernetesLong,
		Run: func(cmd *cobra.Command, args []string) {
			if cmdutil.IsFilenameSliceEmpty(filesOpts.Filenames) {
				handler(cmd, args)
				return
			}

			schema, err := f.Validator(true)
			cmdutil.CheckErr(err)

			cmdNamespace, enforceNamespace, err := f.DefaultNamespace()
			cmdutil.CheckErr(err)

			builder := f.NewUnstructuredBuilder()
			builder.Schema(schema)
			builder.ContinueOnError()
			builder.NamespaceParam(cmdNamespace)
			builder.DefaultNamespace()
			builder.FilenameParam(enforceNamespace, &filesOpts)
			builder.Flatten()

			result := builder.Do()
			cmdutil.CheckErr(result.Err())

			err = result.Visit(func(info *resource.Info, err error) error {
				if err != nil {
					return err
				}

				if err := createAndRefresh(info); err != nil {
					return err
				}

				kind, err := info.Mapping.Kind(info.Object)
				if err != nil {
					kind = "<unkn>"
				}

				// we could use cmdutil.PrintSuccess() here, but that has more dependencies which we don't care
				// for for now.
				fmt.Fprintf(out, "Hello %s %s\n", kind, info.Name)

				return nil
			})
			cmdutil.CheckErr(err)
		},
	}

	cmdutil.AddFilenameOptionFlags(cmd, &filesOpts, "File describing resource to create")
	cmd.MarkFlagRequired("filename")

	return cmd
}
