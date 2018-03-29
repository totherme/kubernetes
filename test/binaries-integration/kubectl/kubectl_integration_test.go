package kubectl_test

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

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
					ExpectStdoutTo(haveImportantLabels).Succeeds()

				// kube::test::describe_object_events_assert pods 'valid-pod'
				By("printing events information by default")
				kubeCtl.WithArgs("describe", "--show-events=true", "pods", "valid-pod").
					ExpectStdoutTo(HaveEvents()).Succeeds()

				// kube::test::describe_object_events_assert pods 'valid-pod' false
				By("not printing events information when show-events=false")
				kubeCtl.WithArgs("describe", "--show-events=false", "pods", "valid-pod").
					ExpectStdoutTo(NotHaveEvents()).Succeeds()

				// kube::test::describe_object_events_assert pods 'valid-pod' true
				By("printing events information when show-events=true")
				kubeCtl.WithArgs("describe", "--show-events=true", "pods", "valid-pod").
					ExpectStdoutTo(HaveEvents()).Succeeds()
			})

			It("succeeds describing resource only", func() {
				// kube::test::describe_resource_assert pods "Name:" "Image:" "Node:" "Labels:" "Status:"
				kubeCtl.WithArgs("describe", "pods").ExpectStdoutTo(haveImportantLabels).Succeeds()
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

	Context("in the default namespace", func() {
		BeforeEach(func() {
			restartControlPlane()
		})
		It("all expected resource APIs are found", func() {
			containSomeStuffButNotTheOther := And(
				ContainAll(
					"/api/v1/namespaces/default/pods 200 OK",
					"/api/v1/namespaces/default/replicationcontrollers 200 OK",
					"/api/v1/namespaces/default/services 200 OK",
					"/apis/apps/v1/namespaces/default/daemonsets 200 OK",
					"/apis/apps/v1/namespaces/default/deployments 200 OK",
					"/apis/apps/v1/namespaces/default/replicasets 200 OK",
					"/apis/apps/v1/namespaces/default/statefulsets 200 OK",
					"/apis/autoscaling/v1/namespaces/default/horizontalpodautoscalers 200",
					"/apis/batch/v1/namespaces/default/jobs 200 OK",
				),
				NotContainAny(
					"/apis/extensions/v1beta1/namespaces/default/daemonsets 200 OK",
					"/apis/extensions/v1beta1/namespaces/default/deployments 200 OK",
					"/apis/extensions/v1beta1/namespaces/default/replicasets 200 OK",
				),
			)

			kubeCtl.WithArgs("--v=6", "--namespace", "default", "get", "all", "--chunk-size=0").
				ExpectStderrTo(containSomeStuffButNotTheOther).Succeeds()
		})
	})

	FContext("with a client", func() {
		XIt("succeeds", func() {
			kubeCtl.
				Run("get", "--raw", "/version")
		})

		Context("when testing client versions", func() {
			var (
				allVersions       string
				allVersionsParsed Versions
			)
			BeforeEach(func() {
				var err error
				allVersions, _ = kubeCtl.Run("version")
				allVersionsParsed, err = parseAllVersions(allVersions)
				Expect(err).NotTo(HaveOccurred())
			})
			Context("with plain output", func() {
				var (
					stdoutWithClientFlag string
					clientVersion        string
					serverVersion        string
				)
				BeforeEach(func() {
					stdoutWithClientFlag, _ = kubeCtl.Run("version", "--client")
					clientVersion = getVersionInfo(allVersions, "Client")
					serverVersion = getVersionInfo(allVersions, "Server")
				})

				It("outputs client information", func() {
					Expect(strings.TrimSpace(stdoutWithClientFlag)).To(Equal(clientVersion))
				})
				It("doesn't output server information", func() {
					Expect(strings.TrimSpace(stdoutWithClientFlag)).NotTo(ContainSubstring(serverVersion))
				})
			})
			Context("with json output for all components", func() {
				var (
					stdoutAsJSONString string
					versions           Versions
					args               []string
				)
				BeforeEach(func() {
					args = []string{"version", "--output", "json"}
				})

				JustBeforeEach(func() {
					stdoutAsJSONString, _ = kubeCtl.Run(args...)
					versions = Versions{}
					Expect(json.Unmarshal([]byte(stdoutAsJSONString), &versions)).To(Succeed())
				})
				It("outputs correct server information", func() {
					Expect(allVersionsParsed.ServerVersion).To(Equal(versions.ServerVersion))
				})
				It("outputs correct client information", func() {
					Expect(allVersionsParsed.ClientVersion).To(Equal(versions.ClientVersion))
				})
				Context("with --client", func() {
					BeforeEach(func() {
						args = []string{"version", "--client", "--output", "json"}
					})
					It("outputs correct client information", func() {
						Expect(allVersionsParsed.ClientVersion).To(Equal(versions.ClientVersion))
					})
					It("outputs no server information", func() {
						Expect(versions.ServerVersion).To(BeNil())
					})
				})
				Context("with --short", func() {
					BeforeEach(func() {
						args = []string{"version", "--short", "--output", "json"}
					})

					It("ignores the --short flag", func() {
						Expect(allVersionsParsed.ClientVersion).To(Equal(versions.ClientVersion))
						Expect(allVersionsParsed.ServerVersion).To(Equal(versions.ServerVersion))
					})
				})
			})
		})
	})
})

func getVersionInfo(cmdOutput, component string) string {
	versionInfo := strings.Split(cmdOutput, "\n")
	for _, line := range versionInfo {
		if strings.HasPrefix(line, fmt.Sprintf("%s Version", component)) {
			return line
		}
	}
	return ""
}

func parseAllVersions(asString string) (Versions, error) {
	v := Versions{}
	versionStrings := strings.Split(asString, "\n")

	for _, s := range versionStrings {
		if s == "" {
			continue
		}

		if strings.HasPrefix(s, "Server Version:") {
			v.ServerVersion = parseVersion(s)
			continue
		}

		if strings.HasPrefix(s, "Client Version:") {
			v.ClientVersion = parseVersion(s)
			continue
		}

		return v, fmt.Errorf("Unkonwn line '%s'", s)
	}
	return v, nil
}

func parseVersion(asString string) *VersionInfo {
	v := VersionInfo{}
	find := func(search string) string {
		re := regexp.MustCompile(search + `:"([^"]*)"`)
		matches := re.FindStringSubmatch(asString)
		if len(matches) < 2 {
			return "<nil>"
		}
		return matches[1]
	}
	v.Major = find("Major")
	v.Minor = find("Minor")
	v.GitVersion = find("GitVersion")
	v.GitCommit = find("GitCommit")
	v.GitTreeState = find("GitTreeState")
	v.BuildDate = find("BuildDate")
	v.GoVersion = find("GoVersion")
	v.Compiler = find("Compiler")
	v.Platform = find("Platform")
	return &v
}

type Versions struct {
	ClientVersion *VersionInfo `json:"clientVersion"`
	ServerVersion *VersionInfo `json:"serverVersion"`
}
type VersionInfo struct {
	Major        string `json:"major"`
	Minor        string `json:"minor"`
	GitVersion   string `json:"gitVersion"`
	GitCommit    string `json:"gitCommit"`
	GitTreeState string `json:"gitTreeState"`
	BuildDate    string `json:"buildDate"`
	GoVersion    string `json:"goVersion"`
	Compiler     string `json:"compiler"`
	Platform     string `json:"platform"`
}
