package cmd

import (
	"bytes"
	"net/http"
	"strings"
	"testing"

	"fmt"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest/fake"
	"k8s.io/kubernetes/pkg/apis/core"
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

func TestHelloKubernetesWorksWithAService(t *testing.T) {
	var b bytes.Buffer

	file := "../../../examples/cluster-dns/dns-backend-service.yaml"
	name := "dns-backend"
	kind := "Service"

	f, _ := setupHelloKubernetesTestFactory(func(req *http.Request, codec runtime.Codec) (*http.Response, error) {
		_, service, _ := testData()
		service.Items[0].Name = name

		switch p, m := req.URL.Path, req.Method; {
		case p == "/namespaces/test/services" && m == http.MethodPost:
			return &http.Response{
				StatusCode: http.StatusCreated,
				Header:     defaultHeader(),
				Body:       objBody(codec, &service.Items[0]),
			}, nil
		default:
			t.Fatalf("unexpected request: %#v\n%#v", req.URL, req)
			return nil, nil
		}
	})

	cmd := NewCmdHelloKubernetes(f, &b, nil, dummyErrorHandler)
	if cmd == nil {
		t.Errorf("Expected NewCmdHelloKubernetes() not to return nil")
		t.FailNow()
	}

	cmd.Flags().Set("filename", file)
	cmd.Run(cmd, []string{})

	actual := b.String()
	expected := fmt.Sprintf("Hello %s %s\n", kind, name)
	if actual != expected {
		t.Errorf("Expected output %s, got: %s", expected, actual)
	}
}

func TestHelloKubernetesWorksWithAReplicationController(t *testing.T) {
	var b bytes.Buffer

	filename := "../../../examples/guestbook/legacy/redis-master-controller.yaml"
	name := "redis-master"
	kind := "ReplicationController"

	f, _ := setupHelloKubernetesTestFactory(func(req *http.Request, codec runtime.Codec) (*http.Response, error) {
		_, _, rc := testData()
		rc.Items[0].Name = name

		switch p, m := req.URL.Path, req.Method; {
		case p == "/namespaces/test/replicationcontrollers" && m == http.MethodPost:
			return &http.Response{
				StatusCode: http.StatusCreated,
				Header:     defaultHeader(),
				Body:       objBody(codec, &rc.Items[0]),
			}, nil
		default:
			t.Fatalf("unexpected request: %#v\n%#v", req.URL, req)
			return nil, nil
		}
	})

	cmd := NewCmdHelloKubernetes(f, &b, nil, dummyErrorHandler)
	if cmd == nil {
		t.Errorf("Expected NewCmdHelloKubernetes() not to return nil")
		t.FailNow()
	}

	cmd.Flags().Set("filename", filename)
	cmd.Run(cmd, []string{})

	actual := b.String()
	expected := fmt.Sprintf("Hello %s %s\n", kind, name)
	if actual != expected {
		t.Errorf("Expected output %s, got: %s", expected, actual)
	}
}

func TestNewCmdHelloKubernetesWorksWithMultipleFiles(t *testing.T) {
	var b bytes.Buffer

	filename1 := "../../../examples/guestbook-go/redis-master-controller.json"
	filename2 := "../../../examples/guestbook/legacy/redis-master-controller.yaml"
	name1 := "test replication controller 1"
	name2 := "test replication controller 2"
	//kind1 := "ReplicationController"
	//kind2 := "ReplicationController"

	testReplicationControllers := &core.ReplicationControllerList{
		ListMeta: metav1.ListMeta{
			ResourceVersion: "17",
		},
		Items: []core.ReplicationController{
			{
				ObjectMeta: metav1.ObjectMeta{Name: name1, Namespace: "test", ResourceVersion: "18"},
				Spec: core.ReplicationControllerSpec{
					Replicas: 1,
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{Name: name2, Namespace: "test", ResourceVersion: "28"},
				Spec: core.ReplicationControllerSpec{
					Replicas: 2,
				},
			},
		},
	}

	whichOneToReturn := -1
	f, _ := setupHelloKubernetesTestFactory(func(req *http.Request, codec runtime.Codec) (*http.Response, error) {
		switch p, m := req.URL.Path, req.Method; {
		case p == "/namespaces/test/replicationcontrollers" && m == http.MethodPost:
			whichOneToReturn += 1
			return &http.Response{
				StatusCode: http.StatusCreated,
				Header:     defaultHeader(),
				Body:       objBody(codec, &testReplicationControllers.Items[whichOneToReturn]),
			}, nil
		default:
			t.Fatalf("unexpected request: %#v\n%#v", req.URL, req)
			return nil, nil
		}
	})

	cmd := NewCmdHelloKubernetes(f, &b, nil, dummyErrorHandler)
	if cmd == nil {
		t.Errorf("Expected NewCmdHelloKubernetes() not to return nil")
		t.FailNow()
	}

	cmd.Flags().Set("filename", filename1)
	cmd.Flags().Set("filename", filename2)
	cmd.Run(cmd, []string{})

	actual := b.String()
	expected := "Hello ReplicationController " + name1 + "\nHello ReplicationController " + name2 + "\n"
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

func dummyErrorHandler(_ *cobra.Command, _ []string) {
}

type responder func(req *http.Request, codec runtime.Codec) (*http.Response, error)

func setupHelloKubernetesTestFactory(resCreator responder) (cmdutil.Factory, fake.RESTClient) {
	f, tf, codec, _ := cmdtesting.NewAPIFactory()

	fakeRestClient := &fake.RESTClient{
		GroupVersion:         schema.GroupVersion{Version: "v1"},
		NegotiatedSerializer: unstructuredSerializer,
		Client: fake.CreateHTTPClient(func(req *http.Request) (*http.Response, error) {
			return resCreator(req, codec)
		}),
	}

	tf.Printer = &testPrinter{}
	tf.UnstructuredClient = fakeRestClient
	tf.Namespace = "test"

	return f, *fakeRestClient
}
