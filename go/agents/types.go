package agents

import (
	"fmt"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/agenttool"
)

func NewRootAgent(m model.LLM, children []agent.Agent) (agent.Agent, error) {
	tools := make([]tool.Tool, 0, len(children))
	for _, child := range children {
		tools = append(tools, agenttool.New(child, nil))
	}

	root, err := llmagent.New(llmagent.Config{
		Name:        "root_agent",
		Model:       m,
		Description: "Main career companion coordinating four specialists (jobs, coding, projects, networking).",
		Instruction: "You are the main Career Companion Agent orchestrating a team of specialists.\n\nYou have four sub-agents:\n- 'job_applications_agent' for job search and applications.\n- 'coding_agent' for coding practice and interview prep.\n- 'networking_agent' for networking and relationship building.\n- 'projects_agent' for personal and portfolio projects.\n\nRouting logic:\n- If the user specifies a hint like [agent:jobs|coding|projects|networking] or 'Option 1/2/3/4', immediately transfer to that child agent without confirmation.\n- Otherwise, infer the most relevant agent and transfer.\n- Never loop or ask the user to pick an option; only ask a brief clarifying question if the request is genuinely ambiguous.\n\nNotes:\n- Read-only policy: do NOT write to the database or ask for data entry.\n- If the user asks to add or update data, route to the relevant specialist; that agent will respond with a JSON write suggestion only.\n- Data lives in the Postgres DB; the UI handles data entry.\n- Your job is routing, not domain analysis.",
		SubAgents:   children,
		Tools:       tools,
	})
	if err != nil {
		return nil, fmt.Errorf("build router: %w", err)
	}
	return root, nil
}
