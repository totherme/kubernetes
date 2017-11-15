package cmd

import (
	"bytes"
	"net/http"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest/fake"
	cmdtesting "k8s.io/kubernetes/pkg/kubectl/cmd/testing"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
)

func TestCmdHelloWorldHasSaneMetaData(t *testing.T) {
	cmd := NewCmdHelloWorld(nil)

	if cmd == nil {
		t.Errorf("Expected NewCmdHelloWorld() not to return nil")
		t.FailNow()
	}

	use := cmd.Use
	if use != "hello-world" {
		t.Errorf("Expected the command to register as 'hello-world', instead registered as '%s'", use)
	}

	if cmd.Short == "" {
		t.Errorf("Expected the command to have a short help text")
	}

	if cmd.Long == "" {
		t.Errorf("Expected the command to have a long help text")
	}
}

func TestCmdHelloWorldCanSayHello(t *testing.T) {
	var b bytes.Buffer

	cmd := NewCmdHelloWorld(&b)

	cmd.Run(nil, nil)

	expected := "Hello World\n"
	actual := b.String()

	if expected != actual {
		t.Errorf("Expected sayHello() to produce '%s', got '%s' instead", expected, actual)
	}
}

func TestCmdHelloKubernetesHasSaneMetaData(t *testing.T) {
	cmd := NewCmdHelloKubernetes(nil, nil, nil, dummyErrorHandler)

	if cmd == nil {
		t.Errorf("Expected NewCmdHelloKubernetes() not to return nil")
		t.FailNow()
	}

	use := cmd.Use
	if !strings.HasPrefix(use, "hello-kubernetes ") {
		t.Errorf("Expected the command to register as 'hello-kubernetes', instead registered as '%s'", use)
	}

	if cmd.Short == "" {
		t.Errorf("Expected the command to have a short help text")
	}

	if cmd.Long == "" {
		t.Errorf("Expected the command to have a long help text")
	}
}

func TestHelloKubernetesWorksWithAJSONFile(t *testing.T) {
	var b bytes.Buffer

	initTestErrorHandler(t)
	f, _ := setupHelloKubernetesTestFactory()

	cmd := NewCmdHelloKubernetes(f, &b, nil, dummyErrorHandler)
	if cmd == nil {
		t.Errorf("Expected NewCmdHelloKubernetes() not to return nil")
		t.FailNow()
	}

	cmd.Flags().Set("filename", "../../../examples/guestbook-go/redis-master-controller.json")
	cmd.Run(cmd, []string{})

	actual := b.String()
	expected := "Hello ReplicationController redis-master\n"
	if actual != expected {
		t.Errorf("Expected output %s, got: %s", expected, actual)
	}
}

func TestHelloKubernetesWorksWithAYAMLFile(t *testing.T) {
	var b bytes.Buffer

	initTestErrorHandler(t)
	f, _ := setupHelloKubernetesTestFactory()

	cmd := NewCmdHelloKubernetes(f, &b, nil, dummyErrorHandler)
	if cmd == nil {
		t.Errorf("Expected NewCmdHelloKubernetes() not to return nil")
		t.FailNow()
	}

	cmd.Flags().Set("filename", "../../../examples/guestbook/legacy/redis-master-controller.yaml")
	cmd.Run(cmd, []string{})

	actual := b.String()
	expected := "Hello ReplicationController redis-master\n"
	if actual != expected {
		t.Errorf("Expected output %s, got: %s", expected, actual)
	}
}

func TestNewCmdHelloKubernetesWorksWithMultipleFiles(t *testing.T) {
	var b bytes.Buffer

	initTestErrorHandler(t)
	f, _ := setupHelloKubernetesTestFactory()

	cmd := NewCmdHelloKubernetes(f, &b, nil, dummyErrorHandler)
	if cmd == nil {
		t.Errorf("Expected NewCmdHelloKubernetes() not to return nil")
		t.FailNow()
	}

	cmd.Flags().Set("filename", "../../../examples/guestbook-go/redis-master-controller.json")
	cmd.Flags().Set("filename", "../../../examples/guestbook/legacy/redis-master-controller.yaml")
	cmd.Run(cmd, []string{})

	actual := b.String()
	expected := "Hello ReplicationController redis-master\nHello ReplicationController redis-master\n"
	if actual != expected {
		t.Errorf("Expected output %s, got: %s", expected, actual)
	}
}

func TestNewCmdHelloKubernetesGracefullyFailsWithNoFiles(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	var handlerCallCount int

	spyHandler := func(_ *cobra.Command, _ []string) {
		handlerCallCount += 1
	}

	f, _, _, _ := cmdtesting.NewAPIFactory()
	cmd := NewCmdHelloKubernetes(f, &stdout, &stderr, spyHandler)
	if cmd == nil {
		t.Errorf("Expected NewCmdHelloKubernetes() not to return nil")
		t.FailNow()
	}

	cmd.Run(cmd, []string{})
	if handlerCallCount != 1 {
		t.Errorf("Expected HelloKubernetes to call the error handler when no arguments are supplied")
	}
}

func TestNewCmdHelloKubernetesCallsTheServerWithTheFile(t *testing.T) {
	initTestErrorHandler(t)

	f, fakeRestClient := setupHelloKubernetesTestFactory()

	buf := bytes.NewBuffer([]byte{})
	errBuf := bytes.NewBuffer([]byte{})

	cmd := NewCmdHelloKubernetes(f, buf, errBuf, dummyErrorHandler)
	cmd.Flags().Set("filename", "../../../examples/guestbook/legacy/redis-master-controller.yaml")

	cmd.Run(cmd, []string{})

	if &fakeRestClient.Req == nil {
		t.Errorf("Expected rest client to be used")
	}
}

func setupHelloKubernetesTestFactory() (cmdutil.Factory, fake.RESTClient) {
	_, _, rc := testData()
	rc.Items[0].Name = "redis-master-controller"

	f, tf, codec, _ := cmdtesting.NewAPIFactory()

	fakeRestClient := &fake.RESTClient{
		GroupVersion:         schema.GroupVersion{Version: "v1"},
		NegotiatedSerializer: unstructuredSerializer,
		Client: fake.CreateHTTPClient(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusCreated,
				Header:     defaultHeader(),
				Body:       objBody(codec, &rc.Items[0]),
			}, nil
		}),
	}

	tf.Printer = &testPrinter{}
	tf.UnstructuredClient = fakeRestClient
	tf.Namespace = "test"

	return f, *fakeRestClient
}

func dummyErrorHandler(_ *cobra.Command, _ []string) {
}
