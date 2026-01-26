package agenttests

import (
	"testing"

	"career-koala/agents"
)

func TestJobAgentMockResponse(t *testing.T) {
	seed := loadSeedData(t)
	reply := "mock jobs response: " + seed.Jobs[0].Title + " at " + seed.Jobs[0].Company
	model := fakeLLM{reply: reply}
	jobAgent, err := agents.NewJobAgent(model, nil)
	if err != nil {
		t.Fatalf("new job agent: %v", err)
	}

	replies := runAgent(t, jobAgent, "Analyze my job applications.")
	if replies[0] != reply {
		t.Fatalf("unexpected reply: %q", replies[0])
	}
}
