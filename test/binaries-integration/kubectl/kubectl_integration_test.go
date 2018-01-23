package kubectl_test

import (
	"fmt"
	"io"
	"io/ioutil"

	"github.com/kubernetes-sig-testing/frameworks/integration"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("KubectlIntegration", func() {
	var (
		kubeCtl *testKubeCtl
		cp      *integration.ControlPlane
	)
	BeforeEach(func() {
		// admissionPluginsEnabled := "Initializers,LimitRanger,ResourceQuota"
		// admissionPluginsDisabled := "ServiceAccount,PersistentVolumeLabel,DefaultStorageClass,DefaultTolerationSeconds,MutatingAdmissionWebhook,ValidatingAdmissionWebhook"
		admissionPluginsEnabled := ""
		admissionPluginsDisabled := "ServiceAccount"

		cp = &integration.ControlPlane{}
		cp.APIServer = &integration.APIServer{
			Path: apiServerPath,
			Args: []string{
				// This will get a bit nices as soon as
				// https://github.com/kubernetes-sigs/testing_frameworks/pull/41 is
				// merged
				"--etcd-servers={{ if .EtcdURL }}{{ .EtcdURL.String }}{{ end }}",
				"--cert-dir={{ .CertDir }}",
				"--insecure-port={{ if .URL }}{{ .URL.Port }}{{ end }}",
				"--insecure-bind-address={{ if .URL }}{{ .URL.Hostname }}{{ end }}",
				"--secure-port=0",
				"--enable-admission-plugins=" + admissionPluginsEnabled,
				"--disable-admission-plugins=" + admissionPluginsDisabled,
			},
		}
		cp.Etcd = &integration.Etcd{Path: etcdPath}

		Expect(cp.Start()).To(Succeed())

		k := cp.KubeCtl()
		k.Path = kubeCtlPath
		kubeCtl = &testKubeCtl{kubeCtl: k}
	})
	AfterEach(func() {
		Expect(cp.Stop()).To(Succeed())
	})

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

		Context("with a go-template specified", func() {
			// kube::test::get_object_assert pods '{{range.items}}{{$id_field}}:{{end}}' 'valid-pod:'
			It("'get pods' succeeds", func() {
				Expect(kubeCtl.RunGoTmpl("{{range.items}}{{.metadata.name}}:{{end}}", "get", "pods")).
					To(Equal("valid-pod:"))
			})
			// kube::test::get_object_assert 'pod valid-pod' '{{$id_field}}' 'valid-pod'
			It("'get pod valid/pod' succeeds", func() {
				Expect(kubeCtl.RunGoTmpl("{{.metadata.name}}", "get", "pod", "valid-pod")).
					To(Equal("valid-pod"))
			})
			// kube::test::get_object_assert 'pod/valid-pod' '{{$id_field}}' 'valid-pod'
			It("'get pod/valid-pod' succeeds", func() {
				Expect(kubeCtl.RunGoTmpl("{{.metadata.name}}", "get", "pod/valid-pod")).
					To(Equal("valid-pod"))
			})
			// kube::test::get_object_assert 'pods/valid-pod' '{{$id_field}}' 'valid-pod'
			It("'get pods/valid-pod' succeeds", func() {
				Expect(kubeCtl.RunGoTmpl("{{.metadata.name}}", "get", "pods/valid-pod")).
					To(Equal("valid-pod"))
			})
		})

		Context("with a jsonpath template specified", func() {
			// kube::test::get_object_jsonpath_assert pods "{.items[*]$id_field}" 'valid-pod'
			It("'get pods' succeeds", func() {
				Expect(kubeCtl.RunJsonPathTmpl("{.items[*].metadata.name}", "get", "pods")).
					To(Equal("valid-pod"))
			})
			// kube::test::get_object_jsonpath_assert 'pod valid-pod' "{$id_field}" 'valid-pod'
			// kube::test::get_object_jsonpath_assert 'pod/valid-pod' "{$id_field}" 'valid-pod'
			// kube::test::get_object_jsonpath_assert 'pods/valid-pod' "{$id_field}" 'valid-pod'
		})
	})
})

type templateType string

const (
	goTemplate       templateType = "go-template"
	jsonPathTemplate templateType = "jsonpath"
)

type testKubeCtl struct {
	kubeCtl  *integration.KubeCtl
	template string
}

func (k *testKubeCtl) Run(args ...string) (string, string) {
	callArgs := []string{}
	callArgs = append(callArgs, args...)
	if k.template != "" {
		callArgs = append(callArgs, "-o", k.template)
	}

	stdout, stderr, err := k.kubeCtl.Run(callArgs...)
	Expect(err).NotTo(HaveOccurred(), "Stdout: %s\nStderr: %s", stdout, stderr)
	return readToString(stdout), readToString(stderr)
}

func (k *testKubeCtl) WithTemplate(ttype templateType, tmpl string) *testKubeCtl {
	clone := &testKubeCtl{
		kubeCtl:  k.kubeCtl,
		template: fmt.Sprintf("%s=%s", ttype, tmpl),
	}
	return clone
}

func (k *testKubeCtl) RunGoTmpl(tmpl string, args ...string) string {
	o, _ := k.WithTemplate(goTemplate, tmpl).Run(args...)
	return o
}

func (k *testKubeCtl) RunJsonPathTmpl(tmpl string, args ...string) string {
	o, _ := k.WithTemplate(jsonPathTemplate, tmpl).Run(args...)
	return o
}

func readToString(r io.Reader) string {
	b, err := ioutil.ReadAll(r)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	return string(b)
}
