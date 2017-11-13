package cmd

import (
	"bytes"
	"testing"
)

func TestCmdHelloWorldHasSaneCommandName(t *testing.T) {
	cmd := NewCmdHelloWorld(nil)
	use := cmd.Use

	if cmd == nil {
		t.Errorf("Expected NewCmdHelloWorld() not to return nil")
	}

	if use != "hello-world" {
		t.Errorf("Expected the command to register as 'hello-world', instead registered as '%s'", use)
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
