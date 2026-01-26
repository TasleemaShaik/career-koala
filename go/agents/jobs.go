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

func NewJobAgent(m model.LLM, dbConn *sql.DB) (agent.Agent, error) {
	listJobs, err := functiontool.New(functiontool.Config{
		Name:        "list_job_applications",
		Description: "List recent job applications (limited set).",
	}, func(ctx tool.Context, limit struct {
		Limit int `json:"limit"`
	}) ([]ckdb.JobApplication, error) {
		return ckdb.ListRecentJobs(ctx, dbConn, limit.Limit)
	})
	if err != nil {
		return nil, err
	}

	return llmagent.New(llmagent.Config{
		Name:        "job_applications_agent",
		Model:       m,
		Description: "Specialist agent that focuses ONLY on job search and applications: resume/cover letter tweaks, tailoring to job descriptions, and creating small daily application tasks.",
		Instruction: "You are the Job Applications Agent.\n- Your responsibility is to help the user make progress on job search tasks using their DB history (the UI handles data entry).\n- Use the 'list_job_applications' tool to fetch recent DB entries (if none exist, say so and give a short starter checklist).\n- Turn those entries into a short, realistic plan for today.\n- Give specific suggestions (for example which type of role/company to target), but keep things achievable.\n- Read-only: do NOT write to the database or request data entry.\n- If the user asks to add/update job applications or goals, return ONLY a JSON write suggestion in a fenced code block using this schema:\n{\n  \"write_requests\": [\n    {\n      \"action\": \"insert\",\n      \"table\": \"job_applications\" | \"daily_goals\" | \"weekly_goals\" | \"monthly_goals\",\n      \"records\": [\n        {\"job_title\":\"\",\"company\":\"\",\"job_link\":\"\",\"applied_date\":\"YYYY-MM-DD\",\"result_date\":null,\"status\":\"applied\",\"notes\":\"\"}\n      ]\n    }\n  ]\n}\n- For goals, use fields: description, target_date|week_of|month_of (YYYY-MM-DD), completed (false), and link IDs (job_application_id, coding_problem_id, project_id, contact_id) as null if unknown.\n- Do NOT handle coding practice, networking, or project planning; those belong to other agents.",
		Tools:       []tool.Tool{listJobs},
	})
}
