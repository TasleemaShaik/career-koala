package agenttests

import (
	"testing"

	"career-koala/agents"
	adkagent "google.golang.org/adk/agent"
)

func TestRootAgentMockResponse(t *testing.T) {
	seed := loadSeedData(t)
	childModel := fakeLLM{reply: "mock child response"}
	jobAgent, err := agents.NewJobAgent(childModel, nil)
	if err != nil {
		t.Fatalf("new job agent: %v", err)
	}
	codingAgent, err := agents.NewCodingAgent(childModel, nil)
	if err != nil {
		t.Fatalf("new coding agent: %v", err)
	}
	projectsAgent, err := agents.NewProjectsAgent(childModel, nil)
	if err != nil {
		t.Fatalf("new projects agent: %v", err)
	}
	networkingAgent, err := agents.NewNetworkingAgent(childModel, nil)
	if err != nil {
		t.Fatalf("new networking agent: %v", err)
	}

	reply := "mock root response: " + seed.Jobs[0].Company + " & " + seed.Coding[0].Title
	rootModel := fakeLLM{reply: reply}
	root, err := agents.NewRootAgent(rootModel, []adkagent.Agent{
		jobAgent,
		codingAgent,
		projectsAgent,
		networkingAgent,
	})
	if err != nil {
		t.Fatalf("new root agent: %v", err)
	}
	if got := len(root.SubAgents()); got != 4 {
		t.Fatalf("expected 4 subagents, got %d", got)
	}

	replies := runAgent(t, root, "Give me a plan.")
	if replies[0] != reply {
		t.Fatalf("unexpected reply: %q", replies[0])
	}
}
