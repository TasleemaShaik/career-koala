package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	migrations "career-koala/migrations"
)

const DefaultPostgresURL = "postgres://postgres@localhost:5433/career_koala?sslmode=disable"

func Init(ctx context.Context) (*sql.DB, error) {
	dsn := DSNFromEnv()
	conn, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	if err := conn.PingContext(ctx); err != nil {
		return nil, err
	}
	if shouldRunMigrations() {
		if err := ensureSchema(ctx, conn); err != nil {
			return nil, err
		}
	}
	return conn, nil
}

func DSNFromEnv() string {
	if dsn := strings.TrimSpace(os.Getenv("DATABASE_URL")); dsn != "" {
		return dsn
	}
	if dsn := strings.TrimSpace(os.Getenv("POSTGRES_DATABASE_URL")); dsn != "" {
		return dsn
	}

	host := envOrDefault("POSTGRES_HOST", "localhost")
	port := envOrDefault("POSTGRES_PORT", "5432")
	user := envOrDefault("POSTGRES_USER", "postgres")
	pass := envOrDefault("POSTGRES_PASSWORD", "postgres")
	dbname := envOrDefault("POSTGRES_DB", "career_koala")
	sslmode := envOrDefault("POSTGRES_SSLMODE", "disable")

	return buildPostgresURL(host, port, user, pass, dbname, sslmode)
}

func envOrDefault(key, fallback string) string {
	val := strings.TrimSpace(os.Getenv(key))
	if val == "" {
		return fallback
	}
	return val
}

func buildPostgresURL(host, port, user, pass, dbname, sslmode string) string {
	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(user, pass),
		Host:   net.JoinHostPort(host, port),
		Path:   dbname,
	}
	if strings.TrimSpace(sslmode) != "" {
		u.RawQuery = url.Values{"sslmode": []string{sslmode}}.Encode()
	}
	return u.String()
}

func shouldRunMigrations() bool {
	val := strings.TrimSpace(strings.ToLower(os.Getenv("RUN_MIGRATIONS")))
	if val == "" {
		return false
	}
	switch val {
	case "1", "true", "yes", "y", "on":
		return true
	default:
		return false
	}
}

func ensureSchema(ctx context.Context, db *sql.DB) error {
	goose.SetBaseFS(migrations.FS)
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	return goose.UpContext(ctx, db, ".")
}

type JobApplication struct {
	ID         int64      `json:"id"`
	JobTitle   string     `json:"job_title"`
	Company    string     `json:"company"`
	JobLink    string     `json:"job_link"`
	Applied    *time.Time `json:"applied_date,omitempty"`
	ResultDate *time.Time `json:"result_date,omitempty"`
	Status     string     `json:"status"`
	Notes      string     `json:"notes"`
}

type CodingProblem struct {
	ID             int64  `json:"id"`
	LeetCodeNumber int    `json:"leetcode_number"`
	Title          string `json:"title"`
	Pattern        string `json:"pattern"`
	ProblemLink    string `json:"problem_link"`
	Difficulty     string `json:"difficulty"`
	AlreadySolved  bool   `json:"already_solved"`
	Notes          string `json:"notes"`
}

type Project struct {
	ID        int64    `json:"id"`
	Name      string   `json:"name"`
	RepoURL   string   `json:"repo_url"`
	Active    bool     `json:"active"`
	TechStack []string `json:"tech_stack"`
	Summary   string   `json:"summary"`
}

type NetworkingContact struct {
	ID                int64  `json:"id"`
	PersonName        string `json:"person_name"`
	HowMet            string `json:"how_met"`
	LinkedInConnected bool   `json:"linkedin_connected"`
	Company           string `json:"company"`
	Position          string `json:"position"`
	Notes             string `json:"notes"`
}

type Goal struct {
	ID             int64     `json:"id"`
	Description    string    `json:"description"`
	TargetDate     time.Time `json:"target_date"`
	Completed      bool      `json:"completed"`
	JobApplication *int64    `json:"job_application_id,omitempty"`
	CodingProblem  *int64    `json:"coding_problem_id,omitempty"`
	Project        *int64    `json:"project_id,omitempty"`
	Contact        *int64    `json:"contact_id,omitempty"`
}

type Meeting struct {
	ID          int64     `json:"id"`
	SessionName string    `json:"session_name"`
	SessionType string    `json:"session_type"`
	SessionTime time.Time `json:"session_time"`
	Location    string    `json:"location"`
	Organizer   string    `json:"organizer"`
	Company     string    `json:"company"`
	Notes       string    `json:"notes"`
}

type Snapshot struct {
	JobApplications    []JobApplication    `json:"job_applications"`
	CodingProblems     []CodingProblem     `json:"coding_problems"`
	Projects           []Project           `json:"projects"`
	NetworkingContacts []NetworkingContact `json:"networking_contacts"`
	DailyGoals         []Goal              `json:"daily_goals"`
	WeeklyGoals        []Goal              `json:"weekly_goals"`
	MonthlyGoals       []Goal              `json:"monthly_goals"`
	Meetings           []Meeting           `json:"meetings"`
}

func InsertJobApplication(ctx context.Context, db *sql.DB, in JobApplication) (int64, error) {
	var id int64
	err := db.QueryRowContext(ctx,
		`INSERT INTO job_applications (job_title, company, job_link, applied_date, result_date, status, notes)
         VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id`,
		in.JobTitle, in.Company, in.JobLink, in.Applied, in.ResultDate, in.Status, in.Notes,
	).Scan(&id)
	return id, err
}

func InsertCodingProblem(ctx context.Context, db *sql.DB, in CodingProblem) (int64, error) {
	var id int64
	err := db.QueryRowContext(ctx,
		`INSERT INTO coding_problems (leetcode_number, title, pattern, problem_link, difficulty, already_solved, notes)
         VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id`,
		in.LeetCodeNumber, in.Title, in.Pattern, in.ProblemLink, in.Difficulty, in.AlreadySolved, in.Notes,
	).Scan(&id)
	return id, err
}

func InsertProject(ctx context.Context, db *sql.DB, in Project) (int64, error) {
	var id int64
	err := db.QueryRowContext(ctx,
		`INSERT INTO projects (name, repo_url, active, tech_stack, summary)
         VALUES ($1,$2,$3,$4,$5) RETURNING id`,
		in.Name, in.RepoURL, in.Active, pqStringArray(in.TechStack), in.Summary,
	).Scan(&id)
	return id, err
}

func InsertNetworkingContact(ctx context.Context, db *sql.DB, in NetworkingContact) (int64, error) {
	var id int64
	err := db.QueryRowContext(ctx,
		`INSERT INTO networking_contacts (person_name, how_met, linkedin_connected, company, position, notes)
         VALUES ($1,$2,$3,$4,$5,$6) RETURNING id`,
		in.PersonName, in.HowMet, in.LinkedInConnected, in.Company, in.Position, in.Notes,
	).Scan(&id)
	return id, err
}

func InsertDailyGoal(ctx context.Context, db *sql.DB, in Goal) (int64, error) {
	var id int64
	err := db.QueryRowContext(ctx,
		`INSERT INTO daily_goals (description, target_date, completed, job_application_id, coding_problem_id, project_id, contact_id)
         VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id`,
		in.Description, in.TargetDate, in.Completed, in.JobApplication, in.CodingProblem, in.Project, in.Contact,
	).Scan(&id)
	return id, err
}

func InsertWeeklyGoal(ctx context.Context, db *sql.DB, in Goal) (int64, error) {
	var id int64
	err := db.QueryRowContext(ctx,
		`INSERT INTO weekly_goals (description, week_of, completed, job_application_id, coding_problem_id, project_id, contact_id)
         VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id`,
		in.Description, in.TargetDate, in.Completed, in.JobApplication, in.CodingProblem, in.Project, in.Contact,
	).Scan(&id)
	return id, err
}

func InsertMonthlyGoal(ctx context.Context, db *sql.DB, in Goal) (int64, error) {
	var id int64
	err := db.QueryRowContext(ctx,
		`INSERT INTO monthly_goals (description, month_of, completed, job_application_id, coding_problem_id, project_id, contact_id)
         VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id`,
		in.Description, in.TargetDate, in.Completed, in.JobApplication, in.CodingProblem, in.Project, in.Contact,
	).Scan(&id)
	return id, err
}

func InsertMeeting(ctx context.Context, db *sql.DB, in Meeting) (int64, error) {
	var id int64
	err := db.QueryRowContext(ctx,
		`INSERT INTO meetings (session_name, session_type, session_time, location, organizer, company, notes)
         VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id`,
		in.SessionName, in.SessionType, in.SessionTime, in.Location, in.Organizer, in.Company, in.Notes,
	).Scan(&id)
	return id, err
}

func ListJobApplications(ctx context.Context, db *sql.DB) ([]JobApplication, error) {
	rows, err := db.QueryContext(ctx, `SELECT id, job_title, COALESCE(company,''), COALESCE(job_link,''), applied_date, result_date, COALESCE(status,''), COALESCE(notes,'') FROM job_applications ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []JobApplication
	for rows.Next() {
		var r JobApplication
		if err := rows.Scan(&r.ID, &r.JobTitle, &r.Company, &r.JobLink, &r.Applied, &r.ResultDate, &r.Status, &r.Notes); err != nil {
			return nil, err
		}
		res = append(res, r)
	}
	return res, rows.Err()
}

func ListCodingProblems(ctx context.Context, db *sql.DB) ([]CodingProblem, error) {
	rows, err := db.QueryContext(ctx, `SELECT id, leetcode_number, title, pattern, problem_link, difficulty, already_solved, notes FROM coding_problems ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []CodingProblem
	for rows.Next() {
		var r CodingProblem
		_ = rows.Scan(&r.ID, &r.LeetCodeNumber, &r.Title, &r.Pattern, &r.ProblemLink, &r.Difficulty, &r.AlreadySolved, &r.Notes)
		res = append(res, r)
	}
	return res, rows.Err()
}

func ListProjects(ctx context.Context, db *sql.DB) ([]Project, error) {
	rows, err := db.QueryContext(ctx, `SELECT id, name, repo_url, active, summary, COALESCE(to_json(tech_stack), '[]'::json) FROM projects ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []Project
	for rows.Next() {
		var r Project
		var techJSON []byte
		if err := rows.Scan(&r.ID, &r.Name, &r.RepoURL, &r.Active, &r.Summary, &techJSON); err != nil {
			return nil, err
		}
		if len(techJSON) > 0 {
			if err := json.Unmarshal(techJSON, &r.TechStack); err != nil {
				return nil, err
			}
		}
		res = append(res, r)
	}
	return res, rows.Err()
}

func ListNetworkingContacts(ctx context.Context, db *sql.DB) ([]NetworkingContact, error) {
	rows, err := db.QueryContext(ctx, `SELECT id, person_name, how_met, linkedin_connected, company, position, notes FROM networking_contacts ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []NetworkingContact
	for rows.Next() {
		var r NetworkingContact
		_ = rows.Scan(&r.ID, &r.PersonName, &r.HowMet, &r.LinkedInConnected, &r.Company, &r.Position, &r.Notes)
		res = append(res, r)
	}
	return res, rows.Err()
}

func ListDailyGoals(ctx context.Context, db *sql.DB) ([]Goal, error) {
	rows, err := db.QueryContext(ctx, `SELECT id, description, target_date, completed, job_application_id, coding_problem_id, project_id, contact_id FROM daily_goals ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []Goal
	for rows.Next() {
		var r Goal
		_ = rows.Scan(&r.ID, &r.Description, &r.TargetDate, &r.Completed, &r.JobApplication, &r.CodingProblem, &r.Project, &r.Contact)
		res = append(res, r)
	}
	return res, rows.Err()
}

func ListWeeklyGoals(ctx context.Context, db *sql.DB) ([]Goal, error) {
	rows, err := db.QueryContext(ctx, `SELECT id, description, week_of, completed, job_application_id, coding_problem_id, project_id, contact_id FROM weekly_goals ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []Goal
	for rows.Next() {
		var r Goal
		_ = rows.Scan(&r.ID, &r.Description, &r.TargetDate, &r.Completed, &r.JobApplication, &r.CodingProblem, &r.Project, &r.Contact)
		res = append(res, r)
	}
	return res, rows.Err()
}

func ListMonthlyGoals(ctx context.Context, db *sql.DB) ([]Goal, error) {
	rows, err := db.QueryContext(ctx, `SELECT id, description, month_of, completed, job_application_id, coding_problem_id, project_id, contact_id FROM monthly_goals ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []Goal
	for rows.Next() {
		var r Goal
		_ = rows.Scan(&r.ID, &r.Description, &r.TargetDate, &r.Completed, &r.JobApplication, &r.CodingProblem, &r.Project, &r.Contact)
		res = append(res, r)
	}
	return res, rows.Err()
}

func ListMeetings(ctx context.Context, db *sql.DB) ([]Meeting, error) {
	rows, err := db.QueryContext(ctx, `SELECT id, session_name, session_type, session_time, location, organizer, company, notes FROM meetings ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []Meeting
	for rows.Next() {
		var r Meeting
		_ = rows.Scan(&r.ID, &r.SessionName, &r.SessionType, &r.SessionTime, &r.Location, &r.Organizer, &r.Company, &r.Notes)
		res = append(res, r)
	}
	return res, rows.Err()
}

func GetSnapshot(ctx context.Context, db *sql.DB) (Snapshot, error) {
	var s Snapshot
	var err error
	if s.JobApplications, err = ListJobApplications(ctx, db); err != nil {
		return s, err
	}
	if s.CodingProblems, err = ListCodingProblems(ctx, db); err != nil {
		return s, err
	}
	if s.Projects, err = ListProjects(ctx, db); err != nil {
		return s, err
	}
	if s.NetworkingContacts, err = ListNetworkingContacts(ctx, db); err != nil {
		return s, err
	}
	if s.DailyGoals, err = ListDailyGoals(ctx, db); err != nil {
		return s, err
	}
	if s.WeeklyGoals, err = ListWeeklyGoals(ctx, db); err != nil {
		return s, err
	}
	if s.MonthlyGoals, err = ListMonthlyGoals(ctx, db); err != nil {
		return s, err
	}
	if s.Meetings, err = ListMeetings(ctx, db); err != nil {
		return s, err
	}
	return s, nil
}

func ListRecentJobs(ctx context.Context, db *sql.DB, limit int) ([]JobApplication, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := db.QueryContext(ctx, `SELECT id, job_title, COALESCE(company,''), COALESCE(job_link,''), applied_date, result_date, COALESCE(status,''), COALESCE(notes,'') FROM job_applications ORDER BY COALESCE(applied_date, result_date) DESC NULLS LAST, id DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []JobApplication
	for rows.Next() {
		var r JobApplication
		if err := rows.Scan(&r.ID, &r.JobTitle, &r.Company, &r.JobLink, &r.Applied, &r.ResultDate, &r.Status, &r.Notes); err != nil {
			return nil, err
		}
		res = append(res, r)
	}
	return res, rows.Err()
}

func ListRecentCoding(ctx context.Context, db *sql.DB, limit int) ([]CodingProblem, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := db.QueryContext(ctx, `SELECT id, leetcode_number, title, pattern, problem_link, difficulty, already_solved, notes FROM coding_problems ORDER BY id DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []CodingProblem
	for rows.Next() {
		var r CodingProblem
		_ = rows.Scan(&r.ID, &r.LeetCodeNumber, &r.Title, &r.Pattern, &r.ProblemLink, &r.Difficulty, &r.AlreadySolved, &r.Notes)
		res = append(res, r)
	}
	return res, rows.Err()
}

func ListRecentProjects(ctx context.Context, db *sql.DB, limit int) ([]Project, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := db.QueryContext(ctx, `SELECT id, name, repo_url, active, summary, COALESCE(to_json(tech_stack), '[]'::json) FROM projects ORDER BY id DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []Project
	for rows.Next() {
		var r Project
		var techJSON []byte
		if err := rows.Scan(&r.ID, &r.Name, &r.RepoURL, &r.Active, &r.Summary, &techJSON); err != nil {
			return nil, err
		}
		if len(techJSON) > 0 {
			if err := json.Unmarshal(techJSON, &r.TechStack); err != nil {
				return nil, err
			}
		}
		res = append(res, r)
	}
	return res, rows.Err()
}

func ListRecentContacts(ctx context.Context, db *sql.DB, limit int) ([]NetworkingContact, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := db.QueryContext(ctx, `SELECT id, person_name, how_met, linkedin_connected, company, position, notes FROM networking_contacts ORDER BY id DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []NetworkingContact
	for rows.Next() {
		var r NetworkingContact
		_ = rows.Scan(&r.ID, &r.PersonName, &r.HowMet, &r.LinkedInConnected, &r.Company, &r.Position, &r.Notes)
		res = append(res, r)
	}
	return res, rows.Err()
}

func pqStringArray(in []string) interface{} {
	if in == nil {
		return []string{}
	}
	return "{" + strings.Join(in, ",") + "}"
}

func UpdateGoalCompleted(ctx context.Context, db *sql.DB, goalType string, id int64, completed bool) error {
	table, err := goalTable(goalType)
	if err != nil {
		return err
	}

	res, err := db.ExecContext(ctx, fmt.Sprintf("UPDATE %s SET completed=$1 WHERE id=$2", table), completed, id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func UpdateJobStatus(ctx context.Context, db *sql.DB, id int64, status string) error {
	status = strings.TrimSpace(status)
	if status == "" {
		return fmt.Errorf("status is required")
	}
	statusLower := strings.ToLower(status)
	isRejected := strings.Contains(statusLower, "reject")
	res, err := db.ExecContext(
		ctx,
		"UPDATE job_applications SET status=$1, result_date=CASE WHEN $2 THEN CURRENT_DATE ELSE result_date END WHERE id=$3",
		status,
		isRejected,
		id,
	)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func UpdateGoal(ctx context.Context, db *sql.DB, goalType string, id int64, completed bool, description string) error {
	table, err := goalTable(goalType)
	if err != nil {
		return err
	}
	desc := strings.TrimSpace(description)
	var res sql.Result
	if desc == "" {
		res, err = db.ExecContext(ctx, fmt.Sprintf("UPDATE %s SET completed=$1 WHERE id=$2", table), completed, id)
	} else {
		res, err = db.ExecContext(ctx, fmt.Sprintf("UPDATE %s SET completed=$1, description=$2 WHERE id=$3", table), completed, desc, id)
	}
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func UpdateGoalCompletedByDescription(ctx context.Context, db *sql.DB, goalType, description string, completed bool) (int64, error) {
	table, err := goalTable(goalType)
	if err != nil {
		return 0, err
	}
	desc := strings.TrimSpace(description)
	if desc == "" {
		return 0, fmt.Errorf("description is required")
	}
	res, err := db.ExecContext(ctx, fmt.Sprintf("UPDATE %s SET completed=$1 WHERE description ILIKE $2", table), completed, desc)
	if err != nil {
		return 0, err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	if rows == 0 {
		return 0, sql.ErrNoRows
	}
	return rows, nil
}

func goalTable(goalType string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(goalType)) {
	case "daily":
		return "daily_goals", nil
	case "weekly":
		return "weekly_goals", nil
	case "monthly":
		return "monthly_goals", nil
	default:
		return "", fmt.Errorf("invalid goal type")
	}
}
