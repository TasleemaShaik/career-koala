package agents

import (
	"database/sql"

	ckdb "career-koala/db"
	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"
)

func NewProjectsAgent(m model.LLM, dbConn *sql.DB) (agent.Agent, error) {
	listProjects, err := functiontool.New(functiontool.Config{
		Name:        "list_projects",
		Description: "List recent projects (limited set).",
	}, func(ctx tool.Context, limit struct {
		Limit int `json:"limit"`
	}) ([]ckdb.Project, error) {
		return ckdb.ListRecentProjects(ctx, dbConn, limit.Limit)
	})
	if err != nil {
		return nil, err
	}

	return llmagent.New(llmagent.Config{
		Name:        "projects_agent",
		Model:       m,
		Description: "Specialist for projects: portfolio gaps, tech depth, and next ideas from DB data.",
		Instruction: "You are the Projects Agent.\n- Your responsibility is to analyze the user's projects and highlight portfolio gaps.\n- Use the 'list_projects' tool to fetch recent DB entries before analyzing.\n- Provide concise recommendations: next project ideas, tech depth, and impact.\n- If there are no records, say so and offer a short checklist plus one follow-up question.\n- Read-only: do NOT write to the database or request data entry.\n- If the user asks to add/update projects or goals, return ONLY a JSON write suggestion in a fenced code block using this schema:\n{\n  \"write_requests\": [\n    {\n      \"action\": \"insert\",\n      \"table\": \"projects\" | \"daily_goals\" | \"weekly_goals\" | \"monthly_goals\",\n      \"records\": [\n        {\"name\":\"\",\"repo_url\":\"\",\"active\":false,\"tech_stack\":[],\"summary\":\"\"}\n      ]\n    }\n  ]\n}\n- For goals, use fields: description, target_date|week_of|month_of (YYYY-MM-DD), completed (false), and link IDs (job_application_id, coding_problem_id, project_id, contact_id) as null if unknown.\n- Do NOT do full data entry; the UI handles that.",
		Tools:       []tool.Tool{listProjects},
	})
}
