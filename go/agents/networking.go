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

func NewNetworkingAgent(m model.LLM, dbConn *sql.DB) (agent.Agent, error) {
	listContacts, err := functiontool.New(functiontool.Config{
		Name:        "list_contacts",
		Description: "List recent networking contacts (limited set).",
	}, func(ctx tool.Context, limit struct {
		Limit int `json:"limit"`
	}) ([]ckdb.NetworkingContact, error) {
		return ckdb.ListRecentContacts(ctx, dbConn, limit.Limit)
	})
	if err != nil {
		return nil, err
	}

	return llmagent.New(llmagent.Config{
		Name:        "networking_agent",
		Model:       m,
		Description: "Specialist agent for networking and relationship building: LinkedIn outreach, recruiter follow-ups, and engagement on posts.",
		Instruction: "You are the Networking Agent.\n- Your responsibility is to help the user build and maintain professional relationships using their DB history (the UI handles data entry).\n- Use the 'list_contacts' tool to fetch recent DB entries (if none exist, say so and give a short starter plan).\n- Turn those into a small set of concrete, non-spammy actions for today.\n- Help the user think of what to say in a personalized, respectful way.\n- Read-only: do NOT write to the database or request data entry.\n- If the user asks to add/update contacts or goals, return ONLY a JSON write suggestion in a fenced code block using this schema:\n{\n  \"write_requests\": [\n    {\n      \"action\": \"insert\",\n      \"table\": \"networking_contacts\" | \"daily_goals\" | \"weekly_goals\" | \"monthly_goals\",\n      \"records\": [\n        {\"person_name\":\"\",\"how_met\":\"\",\"linkedin_connected\":false,\"company\":\"\",\"position\":\"\",\"notes\":\"\"}\n      ]\n    }\n  ]\n}\n- For goals, use fields: description, target_date|week_of|month_of (YYYY-MM-DD), completed (false), and link IDs (job_application_id, coding_problem_id, project_id, contact_id) as null if unknown.\n- Do NOT handle coding practice, deep project work, or resume tailoring.",
		Tools:       []tool.Tool{listContacts},
	})
}
