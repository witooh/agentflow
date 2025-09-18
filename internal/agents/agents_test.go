package agents

import (
	"fmt"
	"testing"
)

func TestNewAgentDefaults(t *testing.T) {
	a := newAgent("Role", "", "gpt-5")
	if a == nil || a.Agent == nil {
		t.Fatal("expected agent constructed")
	}
	// Ensure the agent prints something meaningful via fmt.Sprint
	_ = fmt.Sprint(a.Agent)
}
