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

func NewCodingAgent(m model.LLM, dbConn *sql.DB) (agent.Agent, error) {
	listCoding, err := functiontool.New(functiontool.Config{
		Name:        "list_coding_problems",
		Description: "List recent coding problems (limited set).",
	}, func(ctx tool.Context, limit struct {
		Limit int `json:"limit"`
	}) ([]ckdb.CodingProblem, error) {
		return ckdb.ListRecentCoding(ctx, dbConn, limit.Limit)
	})
	if err != nil {
		return nil, err
	}

	return llmagent.New(llmagent.Config{
		Name:        "coding_agent",
		Model:       m,
		Description: "Specialist agent for coding practice and interview prep: LeetCode-style problems, CS fundamentals, and daily coding habits.",
		Instruction: "You are the Coding Practice Agent.\n- Your responsibility is to help the user plan coding and interview prep using their DB history (the UI handles data entry).\n- Use the 'list_coding_problems' tool to fetch recent DB entries (if none exist, say so and give a short starter checklist).\n- Turn those entries into a structured plan, possibly with problem categories like arrays, graphs, or DP.\n- Encourage consistent, focused practice instead of huge unrealistic goals.\n- Read-only: do NOT write to the database or request data entry.\n- If the user asks to add/update coding problems or goals, return ONLY a JSON write suggestion in a fenced code block using this schema:\n{\n  \"write_requests\": [\n    {\n      \"action\": \"insert\",\n      \"table\": \"coding_problems\" | \"daily_goals\" | \"weekly_goals\" | \"monthly_goals\",\n      \"records\": [\n        {\"leetcode_number\":0,\"title\":\"\",\"pattern\":\"\",\"problem_link\":\"\",\"difficulty\":\"\",\"already_solved\":false,\"notes\":\"\"}\n      ]\n    }\n  ]\n}\n- For goals, use fields: description, target_date|week_of|month_of (YYYY-MM-DD), completed (false), and link IDs (job_application_id, coding_problem_id, project_id, contact_id) as null if unknown.\n- Do NOT handle job applications, networking, or long-term project planning.",
		Tools:       []tool.Tool{listCoding},
	})
}
