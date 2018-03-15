package kubectl_test

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/kubernetes-sig-testing/frameworks/integration"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

func TestKubectl(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Kubectl Suite")
}

var (
	cp      *integration.ControlPlane
	kubeCtl *testKubeCtl
)

var _ = BeforeSuite(func() {
	startControlPlane()
})

var _ = AfterSuite(func() {
	stopControlPlane()
})

func startControlPlane() {
	// admissionPluginsEnabled := "Initializers,LimitRanger,ResourceQuota"
	// admissionPluginsDisabled := "ServiceAccount,PersistentVolumeLabel,DefaultStorageClass,DefaultTolerationSeconds,MutatingAdmissionWebhook,ValidatingAdmissionWebhook"
	admissionPluginsEnabled := ""
	admissionPluginsDisabled := "ServiceAccount"

	cp = &integration.ControlPlane{}
	cp.APIServer = &integration.APIServer{
		Path: getK8sPath("kube-apiserver"),
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
	cp.Etcd = &integration.Etcd{Path: getEtcdPath()}

	Expect(cp.Start()).To(Succeed())

	orgKubeCtl := cp.KubeCtl()
	orgKubeCtl.Path = getK8sPath("kubectl")
	kubeCtl = &testKubeCtl{kubeCtl: orgKubeCtl}
}

func stopControlPlane() {
	Expect(cp.Stop()).To(Succeed())
}

func restartControlPlane() {
	stopControlPlane()
	startControlPlane()
}

func getK8sPath(name string) string {
	return resolveToExecutable(
		filepath.Join(getKubeRoot(), "_output", "bin", name),
		fmt.Sprintf("Have you run `make WHAT=\"cmd/%s\"`?", name),
	)
}

func getEtcdPath() string {
	return resolveToExecutable(
		filepath.Join(getKubeRoot(), "third_party", "etcd", "etcd"),
		"Have you run `./hack/install-etcd.sh`?",
	)
}

func getKubeRoot() string {
	_, filename, _, ok := runtime.Caller(1)
	Expect(ok).To(BeTrue())
	return cdUp(filepath.Dir(filename), 3)
}

func cdUp(path string, count int) string {
	for i := 0; i < count; i++ {
		path = filepath.Dir(path)
	}
	return path
}

func resolveToExecutable(path, message string) string {
	Expect(path).To(BeAnExistingFile(),
		fmt.Sprintf("Expected to find a binary at '%s'. %s", path, message),
	)

	realBin, err := filepath.EvalSymlinks(path)
	Expect(err).NotTo(HaveOccurred(),
		fmt.Sprintf("Could not find link target for '%s'", path),
	)

	stat, err := os.Stat(realBin)
	Expect(err).NotTo(HaveOccurred(),
		fmt.Sprintf("Could not get permissions for '%s'", realBin),
	)

	isExecutable := ((stat.Mode() | 0111) != 0)
	Expect(isExecutable).To(BeTrue(),
		fmt.Sprintf("'%s' is not executable", realBin),
	)

	return realBin
}

type templateType string

const (
	goTemplate       templateType = "go-template"
	jsonPathTemplate templateType = "jsonpath"
)

type kubeCtlTemplate struct {
	tmplType   templateType
	tmplString string
}

type testKubeCtl struct {
	kubeCtl  *integration.KubeCtl
	template kubeCtlTemplate
}

func (k *testKubeCtl) Run(args ...string) (string, string) {
	callArgs := []string{}
	callArgs = append(callArgs, args...)
	if k.template != (kubeCtlTemplate{}) {
		callArgs = append(
			callArgs,
			"-o", fmt.Sprintf("%s=%s", k.template.tmplType, k.template.tmplString),
		)
	}

	stdout, stderr, err := k.kubeCtl.Run(callArgs...)
	Expect(err).NotTo(HaveOccurred(), "Stdout: %s\nStderr: %s", stdout, stderr)
	return readToString(stdout), readToString(stderr)
}

func (k *testKubeCtl) SetTmpl(tmpl kubeCtlTemplate) *testKubeCtl {
	clone := &testKubeCtl{
		kubeCtl:  k.kubeCtl,
		template: tmpl,
	}
	return clone
}

func (k *testKubeCtl) ExpectOutput(matcher types.GomegaMatcher, args ...string) {
	o, _ := k.Run(args...)
	Expect(o).To(matcher)
}

func GoTmpl(tmpl string) kubeCtlTemplate {
	return kubeCtlTemplate{
		tmplType:   goTemplate,
		tmplString: tmpl,
	}
}

func JsonPathTmpl(tmpl string) kubeCtlTemplate {
	return kubeCtlTemplate{
		tmplType:   jsonPathTemplate,
		tmplString: tmpl,
	}
}

func readToString(r io.Reader) string {
	b, err := ioutil.ReadAll(r)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	return string(b)
}
