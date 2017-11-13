package cmd

import (
	"bytes"
	"testing"
)

func TestCmdHelloWorldHasSaneMetaData(t *testing.T) {
	cmd := NewCmdHelloWorld(nil)

	if cmd == nil {
		t.Errorf("Expected NewCmdHelloWorld() not to return nil")
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
