package kubectl_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("KubectlIntegration", func() {
	It("can run 'get pods'", func() {
		stdout, stderr := kubeCtl.Run("get", "pods")
		Expect(stderr).To(ContainSubstring("No resources found."))
		Expect(stdout).To(BeEmpty())
	})

	Context("with no pod deployed", func() {
		It("'get pods' does not fail", func() {
			// kube::test::get_object_assert pods "{{range.items}}{{$id_field}}:{{end}}" ''
			kubeCtl.
				SetTmpl(GoTmpl("{{range.items}}{{.metadata.name}}:{{end}}")).
				ExpectOutput(Equal(""), "get", "pods")
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
			kubeCtl.ExpectOutput(ContainSubstring(`pod "valid-pod" created`), "create", "-f", specFilePath)
		})

		It("can query that pod", func() {
			kubeCtl.Run("get", "pods", "-o", "json")
		})

		It("can query the pod name via go-template", func() {
			kubeCtl = kubeCtl.SetTmpl(GoTmpl("{{range.items}}{{.metadata.name}}:{{end}}"))

			// kube::test::get_object_assert pods '{{range.items}}{{$id_field}}:{{end}}' 'valid-pod:'
			kubeCtl.ExpectOutput(Equal("valid-pod:"), "get", "pods")

			kubeCtl = kubeCtl.SetTmpl(GoTmpl("{{.metadata.name}}"))

			// kube::test::get_object_assert 'pod valid-pod' '{{$id_field}}' 'valid-pod'
			kubeCtl.ExpectOutput(Equal("valid-pod"), "get", "pod", "valid-pod")

			// kube::test::get_object_assert 'pod/valid-pod' '{{$id_field}}' 'valid-pod'
			kubeCtl.ExpectOutput(Equal("valid-pod"), "get", "pod/valid-pod")

			// kube::test::get_object_assert 'pods/valid-pod' '{{$id_field}}' 'valid-pod'
			kubeCtl.ExpectOutput(Equal("valid-pod"), "get", "pods/valid-pod")
		})

		It("can query the pod name via jsonpath-template", func() {
			kubeCtl = kubeCtl.SetTmpl(JsonPathTmpl("{.items[*].metadata.name}"))

			// kube::test::get_object_jsonpath_assert pods "{.items[*]$id_field}" 'valid-pod'
			kubeCtl.ExpectOutput(Equal("valid-pod"), "get", "pods")

			kubeCtl = kubeCtl.SetTmpl(JsonPathTmpl("{.metadata.name}"))

			// kube::test::get_object_jsonpath_assert 'pod valid-pod' "{$id_field}" 'valid-pod'
			kubeCtl.ExpectOutput(Equal("valid-pod"), "get", "pod", "valid-pod")

			// kube::test::get_object_jsonpath_assert 'pod/valid-pod' "{$id_field}" 'valid-pod'
			kubeCtl.ExpectOutput(Equal("valid-pod"), "get", "pod/valid-pod")

			// kube::test::get_object_jsonpath_assert 'pods/valid-pod' "{$id_field}" 'valid-pod'
			kubeCtl.ExpectOutput(Equal("valid-pod"), "get", "pods/valid-pod")
		})
	})
})
