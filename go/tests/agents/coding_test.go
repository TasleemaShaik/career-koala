package agenttests

import (
	"testing"

	"career-koala/agents"
)

func TestCodingAgentMockResponse(t *testing.T) {
	seed := loadSeedData(t)
	reply := "mock coding response: " + seed.Coding[0].Title + " (" + seed.Coding[0].Pattern + ")"
	model := fakeLLM{reply: reply}
	codingAgent, err := agents.NewCodingAgent(model, nil)
	if err != nil {
		t.Fatalf("new coding agent: %v", err)
	}

	replies := runAgent(t, codingAgent, "Review my coding practice.")
	if replies[0] != reply {
		t.Fatalf("unexpected reply: %q", replies[0])
	}
}
