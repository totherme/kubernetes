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
	Expect(path).To(BeAnExistingFile(), "Expected to find a binary at '%s'. %s", path, message)

	realBin, err := filepath.EvalSymlinks(path)
	Expect(err).NotTo(HaveOccurred(), "Could not find link target for '%s'", path)

	stat, err := os.Stat(realBin)
	Expect(err).NotTo(HaveOccurred(), "Could not get permissions for '%s'", realBin)

	isExecutable := ((stat.Mode() | 0111) != 0)
	Expect(isExecutable).To(BeTrue(), "'%s' is not executable", realBin)

	return realBin
}

type outputFormatType string

const (
	goTemplate outputFormatType = "go-template"
	jsonPath   outputFormatType = "jsonpath"
)

type outputFormat struct {
	format   outputFormatType
	template string
}

type testKubeCtl struct {
	kubeCtl       *integration.KubeCtl
	args          []string
	outputFormat  outputFormat
	stdoutMatcher types.GomegaMatcher
	stderrMatcher types.GomegaMatcher
}

func (k *testKubeCtl) clone() *testKubeCtl {
	return &testKubeCtl{
		kubeCtl:       k.kubeCtl,
		args:          k.args,
		outputFormat:  k.outputFormat,
		stdoutMatcher: k.stdoutMatcher,
		stderrMatcher: k.stderrMatcher,
	}
}

func (k *testKubeCtl) Do() (string, string) {
	return k.Run()
}

func (k *testKubeCtl) WithArgs(args ...string) *testKubeCtl {
	clone := k.clone()
	clone.args = args
	return clone
}

func (k *testKubeCtl) Run(args ...string) (string, string) {
	callArgs := k.args
	callArgs = append(callArgs, args...)

	if k.outputFormat != (outputFormat{}) {
		callArgs = append(
			callArgs,
			"-o", fmt.Sprintf("%s=%s", k.outputFormat.format, k.outputFormat.template),
		)
	}

	o, e, err := k.kubeCtl.Run(callArgs...)
	stdout, stderr := readToString(o), readToString(e)

	ExpectWithOffset(1, err).NotTo(HaveOccurred(), "Stdout: %s\nStderr: %s", stdout, stderr)

	if m := k.stdoutMatcher; m != nil {
		ExpectWithOffset(1, stdout).To(m)
	}
	if m := k.stderrMatcher; m != nil {
		ExpectWithOffset(1, stderr).To(m)
	}

	return stdout, stderr
}

func (k *testKubeCtl) WithFormat(fmt outputFormat) *testKubeCtl {
	clone := k.clone()
	clone.outputFormat = fmt
	return clone
}

func (k *testKubeCtl) ExpectStdoutTo(matcher types.GomegaMatcher) *testKubeCtl {
	clone := k.clone()
	clone.stdoutMatcher = matcher
	return clone
}

func (k *testKubeCtl) ExpectStderrTo(matcher types.GomegaMatcher) *testKubeCtl {
	clone := k.clone()
	clone.stderrMatcher = matcher
	return clone
}

func GoTemplate(tmpl string) outputFormat {
	return outputFormat{
		format:   goTemplate,
		template: tmpl,
	}
}

func JsonPath(tmpl string) outputFormat {
	return outputFormat{
		format:   jsonPath,
		template: tmpl,
	}
}

func readToString(r io.Reader) string {
	b, err := ioutil.ReadAll(r)
	ExpectWithOffset(2, err).NotTo(HaveOccurred())
	return string(b)
}

func ContainAll(wanted ...string) types.GomegaMatcher {
	matchers := make([]types.GomegaMatcher, len(wanted))
	for i, s := range wanted {
		matchers[i] = ContainSubstring(s)
	}
	return And(matchers...)
}

func HaveEvents() types.GomegaMatcher {
	return Or(
		ContainSubstring("No events."),
		ContainSubstring("Events:"),
	)
}

func NotHaveEvents() types.GomegaMatcher {
	return Not(HaveEvents())
}
