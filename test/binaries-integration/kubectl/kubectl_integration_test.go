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
		// kube::test::get_object_assert pods "{{range.items}}{{$id_field}}:{{end}}" ''
		It("'get pods' does not fail", func() {
			Expect(kubeCtl.RunGoTmpl("{{range.items}}{{.metadata.name}}:{{end}}", "get", "pods")).
				To(Equal(""))
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
			stdout, stderr := kubeCtl.Run("create", "-f", specFilePath)

			Expect(stderr).To(BeEmpty())
			Expect(stdout).To(ContainSubstring(`pod "valid-pod" created`))
		})

		It("can query that pod", func() {
			kubeCtl.Run("get", "pods", "-o", "json")
		})

		It("can query the pod name via go-template", func() {
			// kube::test::get_object_assert pods '{{range.items}}{{$id_field}}:{{end}}' 'valid-pod:'
			Expect(kubeCtl.RunGoTmpl("{{range.items}}{{.metadata.name}}:{{end}}", "get", "pods")).
				To(Equal("valid-pod:"))

				// kube::test::get_object_assert 'pod valid-pod' '{{$id_field}}' 'valid-pod'
			Expect(kubeCtl.RunGoTmpl("{{.metadata.name}}", "get", "pod", "valid-pod")).
				To(Equal("valid-pod"))

				// kube::test::get_object_assert 'pod/valid-pod' '{{$id_field}}' 'valid-pod'
			Expect(kubeCtl.RunGoTmpl("{{.metadata.name}}", "get", "pod/valid-pod")).
				To(Equal("valid-pod"))

				// kube::test::get_object_assert 'pods/valid-pod' '{{$id_field}}' 'valid-pod'
			Expect(kubeCtl.RunGoTmpl("{{.metadata.name}}", "get", "pods/valid-pod")).
				To(Equal("valid-pod"))
		})

		It("can query the pod name via jsonpath-template", func() {
			// kube::test::get_object_jsonpath_assert pods "{.items[*]$id_field}" 'valid-pod'
			Expect(kubeCtl.RunJsonPathTmpl("{.items[*].metadata.name}", "get", "pods")).
				To(Equal("valid-pod"))

				// kube::test::get_object_jsonpath_assert 'pod valid-pod' "{$id_field}" 'valid-pod'
			Expect(kubeCtl.RunJsonPathTmpl("{.metadata.name}", "get", "pod", "valid-pod")).
				To(Equal("valid-pod"))

				// kube::test::get_object_jsonpath_assert 'pod/valid-pod' "{$id_field}" 'valid-pod'
			Expect(kubeCtl.RunJsonPathTmpl("{.metadata.name}", "get", "pod/valid-pod")).
				To(Equal("valid-pod"))

				// kube::test::get_object_jsonpath_assert 'pods/valid-pod' "{$id_field}" 'valid-pod'
			Expect(kubeCtl.RunJsonPathTmpl("{.metadata.name}", "get", "pods/valid-pod")).
				To(Equal("valid-pod"))
		})
	})
})
