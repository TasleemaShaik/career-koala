package agenttests

import (
	"testing"

	"career-koala/agents"
)

func TestNetworkingAgentMockResponse(t *testing.T) {
	seed := loadSeedData(t)
	reply := "mock networking response: " + seed.Contacts[0].Name + " at " + seed.Contacts[0].Company
	model := fakeLLM{reply: reply}
	networkingAgent, err := agents.NewNetworkingAgent(model, nil)
	if err != nil {
		t.Fatalf("new networking agent: %v", err)
	}

	replies := runAgent(t, networkingAgent, "Suggest networking follow-ups.")
	if replies[0] != reply {
		t.Fatalf("unexpected reply: %q", replies[0])
	}
}
