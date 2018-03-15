package kubectl_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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
				SetOutputFormat(GoTemplate("{{range.items}}{{.metadata.name}}:{{end}}")).
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

		It("can query that pod", func() {
			kubeCtl.Run("get", "pods", "-o", "json")
		})

		It("can query the pod name via go-template", func() {
			kubeCtl = kubeCtl.
				SetOutputFormat(GoTemplate("{{range.items}}{{.metadata.name}}:{{end}}")).
				ExpectStdoutTo(Equal("valid-pod:"))

			// kube::test::get_object_assert pods '{{range.items}}{{$id_field}}:{{end}}' 'valid-pod:'
			kubeCtl.Run("get", "pods")

			kubeCtl = kubeCtl.
				SetOutputFormat(GoTemplate("{{.metadata.name}}")).
				ExpectStdoutTo(Equal("valid-pod"))

			// kube::test::get_object_assert 'pod valid-pod' '{{$id_field}}' 'valid-pod'
			kubeCtl.Run("get", "pod", "valid-pod")

			// kube::test::get_object_assert 'pod/valid-pod' '{{$id_field}}' 'valid-pod'
			kubeCtl.Run("get", "pod/valid-pod")

			// kube::test::get_object_assert 'pods/valid-pod' '{{$id_field}}' 'valid-pod'
			kubeCtl.Run("get", "pods/valid-pod")
		})

		It("can query the pod name via jsonpath-template", func() {
			kubeCtl = kubeCtl.
				SetOutputFormat(JsonPath("{.items[*].metadata.name}")).
				ExpectStdoutTo(Equal("valid-pod"))

			// kube::test::get_object_jsonpath_assert pods "{.items[*]$id_field}" 'valid-pod'
			kubeCtl.Run("get", "pods")

			kubeCtl = kubeCtl.SetOutputFormat(JsonPath("{.metadata.name}"))

			// kube::test::get_object_jsonpath_assert 'pod valid-pod' "{$id_field}" 'valid-pod'
			kubeCtl.Run("get", "pod", "valid-pod")

			// kube::test::get_object_jsonpath_assert 'pod/valid-pod' "{$id_field}" 'valid-pod'
			kubeCtl.Run("get", "pod/valid-pod")

			// kube::test::get_object_jsonpath_assert 'pods/valid-pod' "{$id_field}" 'valid-pod'
			kubeCtl.Run("get", "pods/valid-pod")
		})
	})
})
