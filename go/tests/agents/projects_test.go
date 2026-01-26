package agenttests

import (
	"strings"
	"testing"

	"career-koala/agents"
)

func TestProjectsAgentMockResponse(t *testing.T) {
	seed := loadSeedData(t)
	tech := strings.Join(seed.Projects[0].TechStack, ",")
	reply := "mock projects response: " + seed.Projects[0].Name + " [" + tech + "]"
	model := fakeLLM{reply: reply}
	projectsAgent, err := agents.NewProjectsAgent(model, nil)
	if err != nil {
		t.Fatalf("new projects agent: %v", err)
	}

	replies := runAgent(t, projectsAgent, "Assess my project portfolio.")
	if replies[0] != reply {
		t.Fatalf("unexpected reply: %q", replies[0])
	}
}
