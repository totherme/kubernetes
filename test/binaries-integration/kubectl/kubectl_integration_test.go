package kubectl_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

var _ = Describe("KubectlIntegration", func() {
	It("can run 'get pods'", func() {
		kubeCtl.
			ExpectStderrTo(ContainSubstring("No resources found.")).
			ExpectStdoutTo(BeEmpty()).
			Run("get", "pods")
	})

	Context("with no pod deployed", func() {
		It("'get pods' does not fail", func() {
			// kube::test::get_object_assert pods "{{range.items}}{{$id_field}}:{{end}}" ''
			kubeCtl.
				WithFormat(GoTemplate("{{range.items}}{{.metadata.name}}:{{end}}")).
				ExpectStdoutTo(Equal("")).
				Run("get", "pods")
		})
	})

	Context("with a valid pod deployed", func() {
		AfterEach(func() {
			restartControlPlane()
		})
		BeforeEach(func() {
			specFilePath := getKubeRoot() + "/test/fixtures/doc-yaml/admin/limitrange/valid-pod.yaml"
			Expect(specFilePath).To(BeARegularFile())

			// kubectl create "${kube_flags[@]}" -f test/fixtures/doc-yaml/admin/limitrange/valid-pod.yaml
			kubeCtl.ExpectStdoutTo(ContainSubstring(`pod "valid-pod" created`)).Run("create", "-f", specFilePath)
		})

		Context("querying pods", func() {
			It("succeeds", func() {
				kubeCtl.Run("get", "pods", "-o", "json")
			})

			Context("using go-template", func() {
				It("succeeds for a list of pods", func() {
					By("Setting the output go-template")
					kubeCtl = kubeCtl.WithFormat(GoTemplate("{{range.items}}{{.metadata.name}}:{{end}}"))

					By("checking the templated output")
					// kube::test::get_object_assert pods '{{range.items}}{{$id_field}}:{{end}}' 'valid-pod:'
					kubeCtl.ExpectStdoutTo(Equal("valid-pod:")).Run("get", "pods")
				})

				It("succeeds for single pods", func() {
					By("Setting the output go-template")
					kubeCtl = kubeCtl.WithFormat(GoTemplate("{{.metadata.name}}"))

					By("checking the templated output")
					// kube::test::get_object_assert 'pod valid-pod' '{{$id_field}}' 'valid-pod'
					kubeCtl.ExpectStdoutTo(Equal("valid-pod")).Run("get", "pod", "valid-pod")
					// kube::test::get_object_assert 'pod/valid-pod' '{{$id_field}}' 'valid-pod'
					kubeCtl.ExpectStdoutTo(Equal("valid-pod")).Run("get", "pod/valid-pod")
					// kube::test::get_object_assert 'pods/valid-pod' '{{$id_field}}' 'valid-pod'
					kubeCtl.ExpectStdoutTo(Equal("valid-pod")).Run("get", "pods/valid-pod")
				})
			})

			Context("using jsonPath", func() {
				It("succeeds for a list of pods", func() {
					By("setting up the jsonPath expression")
					kubeCtl = kubeCtl.WithFormat(JsonPath("{.items[*].metadata.name}"))

					By("checking the templated output")
					// kube::test::get_object_jsonpath_assert pods "{.items[*]$id_field}" 'valid-pod'
					kubeCtl.ExpectStdoutTo(Equal("valid-pod")).Run("get", "pods")
				})

				It("succeeds for single pods", func() {
					By("setting up the jsonPath expression")
					kubeCtl = kubeCtl.WithFormat(JsonPath("{.metadata.name}"))

					By("checking the templated output")
					// kube::test::get_object_jsonpath_assert 'pod valid-pod' "{$id_field}" 'valid-pod'
					kubeCtl.ExpectStdoutTo(Equal("valid-pod")).Run("get", "pod", "valid-pod")
					// kube::test::get_object_jsonpath_assert 'pod/valid-pod' "{$id_field}" 'valid-pod'
					kubeCtl.ExpectStdoutTo(Equal("valid-pod")).Run("get", "pod/valid-pod")
					// kube::test::get_object_jsonpath_assert 'pods/valid-pod' "{$id_field}" 'valid-pod'
					kubeCtl.ExpectStdoutTo(Equal("valid-pod")).Run("get", "pods/valid-pod")
				})
			})
		})

		Context("describing resources", func() {
			var haveImportantLabels types.GomegaMatcher
			BeforeEach(func() {
				haveImportantLabels = ContainAll("Name:", "Image:", "Node:", "Labels:", "Status:")
			})
			It("succeeds", func() {
				// kube::test::describe_object_assert pods 'valid-pod' "Name:" "Image:" "Node:" "Labels:" "Status:"
				By("printing detailed information")
				kubeCtl.WithArgs("describe", "pods", "valid-pod").
					ExpectStdoutTo(haveImportantLabels).Do()

				// kube::test::describe_object_events_assert pods 'valid-pod'
				By("printing events information by default")
				kubeCtl.WithArgs("describe", "--show-events=true", "pods", "valid-pod").
					ExpectStdoutTo(HaveEvents()).Do()

				// kube::test::describe_object_events_assert pods 'valid-pod' false
				By("not printing events information when show-events=false")
				kubeCtl.WithArgs("describe", "--show-events=false", "pods", "valid-pod").
					ExpectStdoutTo(NotHaveEvents()).Do()

				// kube::test::describe_object_events_assert pods 'valid-pod' true
				By("printing events information when show-events=true")
				kubeCtl.WithArgs("describe", "--show-events=true", "pods", "valid-pod").
					ExpectStdoutTo(HaveEvents()).Do()
			})

			It("succeeds describing resource only", func() {
				// kube::test::describe_resource_assert pods "Name:" "Image:" "Node:" "Labels:" "Status:"
				kubeCtl.WithArgs("describe", "pods").ExpectStdoutTo(haveImportantLabels).Do()
			})
		})
	})

	Context("namespace configured", func() {
		It("succceeds", func() {
			// kube::test::get_object_assert 'namespaces' '{{range.items}}{{ if eq $id_field \"test-kubectl-describe-pod\" }}found{{end}}{{end}}:' ':'
			By("making sure describing a non-existant namespace won't fail")
			kubeCtl.WithArgs("get", "namespaces").
				WithFormat(GoTemplate(`{{range.items}}{{ if eq .metadata.name "test-kubectl-describe-pod" }}found{{end}}{{end}}:`)).
				ExpectStdoutTo(Equal(":")).Succeeds()

			// kubectl create namespace test-kubectl-describe-pod
			By("creating a namespace")
			kubeCtl.WithArgs("create", "namespace", "test-kubectl-describe-pod").
				Succeeds()

			// kube::test::get_object_assert 'namespaces/test-kubectl-describe-pod' "{{$id_field}}" 'test-kubectl-describe-pod'
			By("enuring the namespace exists now")
			kubeCtl.WithArgs("get", "namespace", "test-kubectl-describe-pod").
				WithFormat(GoTemplate("{{.metadata.name}}")).Succeeds()
		})
	})
})
