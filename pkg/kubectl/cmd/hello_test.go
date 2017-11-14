package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
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
	cmd := NewCmdHelloKubernetes(nil, nil, dummyErrorHandler)

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

	cmd := NewCmdHelloKubernetes(&b, nil, dummyErrorHandler)
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

	cmd := NewCmdHelloKubernetes(&b, nil, dummyErrorHandler)
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

func TestNewCmdHelloKubernetesGracefullyFailsWithNoFiles(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	var handlerCallCount int

	spyHandler := func(_ *cobra.Command, _ []string) {
		handlerCallCount += 1
	}

	cmd := NewCmdHelloKubernetes(&stdout, &stderr, spyHandler)
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
