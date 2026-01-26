package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type WriteRequest struct {
	Action  string                   `json:"action"`
	Table   string                   `json:"table"`
	Records []map[string]interface{} `json:"records"`
}

type WritePayload struct {
	WriteRequests []WriteRequest `json:"write_requests"`
}

var writeJSONBlockRegex = regexp.MustCompile("(?s)```(?:json)?\\s*(\\{.*?\\})\\s*```")

func ExtractWritePayload(text string) (*WritePayload, string, string, error) {
	raw := ""
	if match := writeJSONBlockRegex.FindStringSubmatch(text); match != nil {
		raw = match[1]
	} else {
		trimmed := strings.TrimSpace(text)
		if strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}") {
			raw = trimmed
		}
	}
	if raw == "" {
		return nil, "", "", nil
	}
	decoder := json.NewDecoder(strings.NewReader(raw))
	decoder.UseNumber()
	var payload WritePayload
	if err := decoder.Decode(&payload); err != nil {
		return nil, "", "", err
	}
	if len(payload.WriteRequests) == 0 {
		return nil, "", "", nil
	}
	summaryParts := make([]string, 0, len(payload.WriteRequests))
	for _, req := range payload.WriteRequests {
		action := strings.ToLower(strings.TrimSpace(req.Action))
		table := strings.TrimSpace(req.Table)
		if action != "insert" || table == "" || len(req.Records) == 0 {
			return nil, "", "", fmt.Errorf("invalid write request")
		}
		summaryParts = append(summaryParts, fmt.Sprintf("%s %d -> %s", action, len(req.Records), table))
	}
	return &payload, raw, strings.Join(summaryParts, ", "), nil
}

func ApplyWriteRequests(ctx context.Context, dbConn *sql.DB, payload WritePayload) (string, error) {
	total := 0
	summaries := []string{}
	for _, req := range payload.WriteRequests {
		action := strings.ToLower(strings.TrimSpace(req.Action))
		table := strings.ToLower(strings.TrimSpace(req.Table))
		if action != "insert" {
			return "", fmt.Errorf("unsupported action: %s", req.Action)
		}
		switch table {
		case "job_applications":
			for _, record := range req.Records {
				if _, err := InsertJobApplication(ctx, dbConn, JobApplication{
					JobTitle:   getString(record, "job_title"),
					Company:    getString(record, "company"),
					JobLink:    getString(record, "job_link"),
					Applied:    getDatePtr(record, "applied_date"),
					ResultDate: getDatePtr(record, "result_date"),
					Status:     getString(record, "status"),
					Notes:      getString(record, "notes"),
				}); err != nil {
					return "", err
				}
				total++
			}
			summaries = append(summaries, fmt.Sprintf("%d job_applications", len(req.Records)))
		case "coding_problems":
			for _, record := range req.Records {
				if _, err := InsertCodingProblem(ctx, dbConn, CodingProblem{
					LeetCodeNumber: getInt(record, "leetcode_number"),
					Title:          getString(record, "title"),
					Pattern:        getString(record, "pattern"),
					ProblemLink:    getString(record, "problem_link"),
					Difficulty:     getString(record, "difficulty"),
					AlreadySolved:  getBool(record, "already_solved"),
					Notes:          getString(record, "notes"),
				}); err != nil {
					return "", err
				}
				total++
			}
			summaries = append(summaries, fmt.Sprintf("%d coding_problems", len(req.Records)))
		case "projects":
			for _, record := range req.Records {
				if _, err := InsertProject(ctx, dbConn, Project{
					Name:      getString(record, "name"),
					RepoURL:   getString(record, "repo_url"),
					Active:    getBool(record, "active"),
					TechStack: getStringSlice(record, "tech_stack"),
					Summary:   getString(record, "summary"),
				}); err != nil {
					return "", err
				}
				total++
			}
			summaries = append(summaries, fmt.Sprintf("%d projects", len(req.Records)))
		case "networking_contacts":
			for _, record := range req.Records {
				if _, err := InsertNetworkingContact(ctx, dbConn, NetworkingContact{
					PersonName:        getString(record, "person_name"),
					HowMet:            getString(record, "how_met"),
					LinkedInConnected: getBool(record, "linkedin_connected"),
					Company:           getString(record, "company"),
					Position:          getString(record, "position"),
					Notes:             getString(record, "notes"),
				}); err != nil {
					return "", err
				}
				total++
			}
			summaries = append(summaries, fmt.Sprintf("%d networking_contacts", len(req.Records)))
		case "daily_goals":
			for _, record := range req.Records {
				date := getDate(record, "target_date")
				if date == nil {
					return "", fmt.Errorf("daily_goals requires target_date")
				}
				if _, err := InsertDailyGoal(ctx, dbConn, Goal{
					Description:    getString(record, "description"),
					TargetDate:     *date,
					Completed:      getBool(record, "completed"),
					JobApplication: getIntPtr(record, "job_application_id"),
					CodingProblem:  getIntPtr(record, "coding_problem_id"),
					Project:        getIntPtr(record, "project_id"),
					Contact:        getIntPtr(record, "contact_id"),
				}); err != nil {
					return "", err
				}
				total++
			}
			summaries = append(summaries, fmt.Sprintf("%d daily_goals", len(req.Records)))
		case "weekly_goals":
			for _, record := range req.Records {
				date := getDate(record, "week_of")
				if date == nil {
					return "", fmt.Errorf("weekly_goals requires week_of")
				}
				if _, err := InsertWeeklyGoal(ctx, dbConn, Goal{
					Description:    getString(record, "description"),
					TargetDate:     *date,
					Completed:      getBool(record, "completed"),
					JobApplication: getIntPtr(record, "job_application_id"),
					CodingProblem:  getIntPtr(record, "coding_problem_id"),
					Project:        getIntPtr(record, "project_id"),
					Contact:        getIntPtr(record, "contact_id"),
				}); err != nil {
					return "", err
				}
				total++
			}
			summaries = append(summaries, fmt.Sprintf("%d weekly_goals", len(req.Records)))
		case "monthly_goals":
			for _, record := range req.Records {
				date := getDate(record, "month_of")
				if date == nil {
					return "", fmt.Errorf("monthly_goals requires month_of")
				}
				if _, err := InsertMonthlyGoal(ctx, dbConn, Goal{
					Description:    getString(record, "description"),
					TargetDate:     *date,
					Completed:      getBool(record, "completed"),
					JobApplication: getIntPtr(record, "job_application_id"),
					CodingProblem:  getIntPtr(record, "coding_problem_id"),
					Project:        getIntPtr(record, "project_id"),
					Contact:        getIntPtr(record, "contact_id"),
				}); err != nil {
					return "", err
				}
				total++
			}
			summaries = append(summaries, fmt.Sprintf("%d monthly_goals", len(req.Records)))
		default:
			return "", fmt.Errorf("unsupported table: %s", req.Table)
		}
	}
	return fmt.Sprintf("%d records (%s)", total, strings.Join(summaries, ", ")), nil
}

func getString(record map[string]interface{}, key string) string {
	val, ok := record[key]
	if !ok || val == nil {
		return ""
	}
	switch v := val.(type) {
	case string:
		return v
	case json.Number:
		return v.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

func getBool(record map[string]interface{}, key string) bool {
	val, ok := record[key]
	if !ok || val == nil {
		return false
	}
	switch v := val.(type) {
	case bool:
		return v
	case string:
		return strings.ToLower(v) == "true"
	case json.Number:
		i, err := v.Int64()
		return err == nil && i != 0
	default:
		return false
	}
}

func getInt(record map[string]interface{}, key string) int {
	val, ok := record[key]
	if !ok || val == nil {
		return 0
	}
	switch v := val.(type) {
	case json.Number:
		i, _ := v.Int64()
		return int(i)
	case float64:
		return int(v)
	case int:
		return v
	case string:
		i, _ := strconv.Atoi(v)
		return i
	default:
		return 0
	}
}

func getIntPtr(record map[string]interface{}, key string) *int64 {
	val, ok := record[key]
	if !ok || val == nil {
		return nil
	}
	switch v := val.(type) {
	case json.Number:
		i, err := v.Int64()
		if err != nil {
			return nil
		}
		return &i
	case float64:
		i := int64(v)
		return &i
	case int:
		i := int64(v)
		return &i
	case string:
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil
		}
		return &i
	default:
		return nil
	}
}

func getDate(record map[string]interface{}, key string) *time.Time {
	val := getString(record, key)
	if strings.TrimSpace(val) == "" {
		return nil
	}
	tm, err := time.Parse("2006-01-02", val)
	if err != nil {
		return nil
	}
	return &tm
}

func getDatePtr(record map[string]interface{}, key string) *time.Time {
	return getDate(record, key)
}

func getStringSlice(record map[string]interface{}, key string) []string {
	val, ok := record[key]
	if !ok || val == nil {
		return []string{}
	}
	switch v := val.(type) {
	case []interface{}:
		out := make([]string, 0, len(v))
		for _, item := range v {
			if item == nil {
				continue
			}
			out = append(out, fmt.Sprintf("%v", item))
		}
		return out
	case []string:
		return v
	default:
		return []string{}
	}
}
