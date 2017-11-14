package cmd

import (
	"bytes"
	"testing"
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
	cmd := NewCmdHelloKubernetes(nil)

	if cmd == nil {
		t.Errorf("Expected NewCmdHelloKubernetes() not to return nil")
		t.FailNow()
	}

	use := cmd.Use
	if use != "hello-kubernetes" {
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

	cmd := NewCmdHelloKubernetes(&b)
	if cmd == nil {
		t.Errorf("Expected NewCmdHelloKubernetes() not to return nil")
		t.FailNow()
	}

	cmd.Flags().Set("filename", "../../../examples/guestbook-go/redis-master-controller.json")
	err := cmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("Expected to be able to run the command without error. Got: %s", err)
	}

	actual := b.String()
	expected := "Hello ReplicationController redis-master\n"
	if actual != expected {
		t.Errorf("Expected output %s, got: %s", expected, actual)
	}
}

func TestHelloKubernetesWorksWithAYAMLFile(t *testing.T) {
	var b bytes.Buffer

	cmd := NewCmdHelloKubernetes(&b)
	if cmd == nil {
		t.Errorf("Expected NewCmdHelloKubernetes() not to return nil")
		t.FailNow()
	}

	cmd.Flags().Set("filename", "../../../examples/guestbook/legacy/redis-master-controller.yaml")
	err := cmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("Expected to be able to run the command without error. Got: %s", err)
	}

	actual := b.String()
	expected := "Hello ReplicationController redis-master\n"
	if actual != expected {
		t.Errorf("Expected output %s, got: %s", expected, actual)
	}
}
