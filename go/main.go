package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"career-koala/agents"
	ckdb "career-koala/db"

	"golang.org/x/oauth2/google"
	"google.golang.org/adk/agent"
	"google.golang.org/adk/model/gemini"
	"google.golang.org/adk/runner"
	"google.golang.org/adk/session"
	"google.golang.org/genai"
)

func main() {
	ctx := context.Background()

	enableAI := boolFromEnv("ENABLE_AI", false)

	conn, err := ckdb.Init(ctx)
	if err != nil {
		log.Fatalf("db init (postgres): %v", err)
	}
	defer conn.Close()

	var sessSvc session.Service
	var runners map[string]*runner.Runner
	if enableAI {
		// Allow overriding the model via env (MODEL_NAME); validate and normalize.
		modelName := os.Getenv("MODEL_NAME")
		project := os.Getenv("GOOGLE_CLOUD_PROJECT")
		location := os.Getenv("VERTEX_LOCATION")
		fmt.Println("Project:", project, "Location:", location, "Model:", modelName)
		if project == "" {
			project = "PROJECT_ID"
		}
		if location == "" {
			location = "us-central1"
		}
		modelName, err = resolveModelName(modelName, project, location)
		if err != nil {
			log.Fatalf("model name: %v", err)
		}
		fmt.Println("Resolved model name:", modelName)
		if boolFromEnv("VALIDATE_MODEL", true) {
			vctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()
			if err := validateModelExists(vctx, modelName); err != nil {
				log.Fatalf("model validation failed: %v", err)
			}
		}
		cfg := &genai.ClientConfig{
			Backend:  genai.BackendVertexAI,
			Project:  project,
			Location: location,
		}
		if err := cfg.UseDefaultCredentials(); err != nil {
			log.Fatalf("credentials: %v", err)
		}
		model, err := gemini.NewModel(ctx, modelName, cfg)
		if err != nil {
			log.Fatalf("create model: %v", err)
		}
		fmt.Println("After gemini.NewModel model:", model)
		jobAgent, err := agents.NewJobAgent(model, conn)
		if err != nil {
			log.Fatalf("job agent: %v", err)
		}
		codingAgent, err := agents.NewCodingAgent(model, conn)
		if err != nil {
			log.Fatalf("coding agent: %v", err)
		}
		projectAgent, err := agents.NewProjectsAgent(model, conn)
		if err != nil {
			log.Fatalf("projects agent: %v", err)
		}
		networkingAgent, err := agents.NewNetworkingAgent(model, conn)
		if err != nil {
			log.Fatalf("networking agent: %v", err)
		}
		root, err := agents.NewRootAgent(model, []agent.Agent{jobAgent, codingAgent, projectAgent, networkingAgent})
		if err != nil {
			log.Fatalf("root agent: %v", err)
		}

		sessSvc = session.InMemoryService()
		rootRunner, err := runner.New(runner.Config{
			AppName:        "career_koala",
			Agent:          root,
			SessionService: sessSvc,
		})
		if err != nil {
			log.Fatalf("runner: %v", err)
		}
		jobRunner, err := runner.New(runner.Config{
			AppName:        "career_koala",
			Agent:          jobAgent,
			SessionService: sessSvc,
		})
		if err != nil {
			log.Fatalf("runner: %v", err)
		}
		codingRunner, err := runner.New(runner.Config{
			AppName:        "career_koala",
			Agent:          codingAgent,
			SessionService: sessSvc,
		})
		if err != nil {
			log.Fatalf("runner: %v", err)
		}
		projectsRunner, err := runner.New(runner.Config{
			AppName:        "career_koala",
			Agent:          projectAgent,
			SessionService: sessSvc,
		})
		if err != nil {
			log.Fatalf("runner: %v", err)
		}
		networkingRunner, err := runner.New(runner.Config{
			AppName:        "career_koala",
			Agent:          networkingAgent,
			SessionService: sessSvc,
		})
		if err != nil {
			log.Fatalf("runner: %v", err)
		}

		runners = map[string]*runner.Runner{
			"root":       rootRunner,
			"jobs":       jobRunner,
			"coding":     codingRunner,
			"projects":   projectsRunner,
			"networking": networkingRunner,
		}
	}

	mux := http.NewServeMux()
	// API only; Next.js UI runs separately.
	if enableAI {
		mux.HandleFunc("/chat", chatHandler(runners, sessSvc, conn))
	} else {
		mux.HandleFunc("/chat", chatDisabledHandler())
	}
	mux.HandleFunc("/meta", metaHandler(enableAI))
	mux.HandleFunc("/data", dataHandler(conn))
	mux.HandleFunc("/jobs", jobCreateHandler(conn))
	mux.HandleFunc("/jobs/status", jobStatusUpdateHandler(conn))
	mux.HandleFunc("/coding", codingCreateHandler(conn))
	mux.HandleFunc("/projects", projectCreateHandler(conn))
	mux.HandleFunc("/networking", networkingCreateHandler(conn))
	mux.HandleFunc("/goals", goalUpdateHandler(conn))

	addr := ":8080"
	log.Printf("CareerKoala API listening on %s (ai=%t).", addr, enableAI)
	log.Fatal(http.ListenAndServe(addr, withCORS(mux)))
}

type chatRequest struct {
	UserID    string `json:"user_id"`
	SessionID string `json:"session_id"`
	Message   string `json:"message"`
	Agent     string `json:"agent"`
}

type chatResponse struct {
	SessionID string   `json:"session_id"`
	Replies   []string `json:"replies"`
	Error     string   `json:"error,omitempty"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func chatHandler(runners map[string]*runner.Runner, sessSvc session.Service, dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var req chatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}
		if req.Message == "" {
			writeError(w, http.StatusBadRequest, "message is required")
			return
		}
		req.Agent = strings.ToLower(strings.TrimSpace(req.Agent))
		selectedRunner := runners["root"]
		if req.Agent != "" && req.Agent != "auto" {
			if alt, ok := runners[req.Agent]; ok {
				selectedRunner = alt
			}
		}
		if req.Agent == "" || req.Agent == "auto" {
			req.Message = normalizeAgentHint(req.Message)
		}
		log.Printf("chat request user=%s session=%s agent=%s msg=%q", req.UserID, req.SessionID, req.Agent, req.Message)
		if req.UserID == "" {
			req.UserID = "demo_user"
		}

		if req.SessionID == "" {
			ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
			defer cancel()
			createResp, err := sessSvc.Create(ctx, &session.CreateRequest{
				AppName: "career_koala",
				UserID:  req.UserID,
			})
			if err != nil {
				writeError(w, http.StatusInternalServerError, "failed to create session")
				return
			}
			req.SessionID = createResp.Session.ID()
		} else {
			ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
			defer cancel()
			_, err := sessSvc.Get(ctx, &session.GetRequest{
				AppName:   "career_koala",
				UserID:    req.UserID,
				SessionID: req.SessionID,
			})
			if err != nil {
				createResp, cerr := sessSvc.Create(ctx, &session.CreateRequest{
					AppName:   "career_koala",
					UserID:    req.UserID,
					SessionID: req.SessionID,
				})
				if cerr != nil {
					writeError(w, http.StatusInternalServerError, "failed to create session")
					return
				}
				req.SessionID = createResp.Session.ID()
			}
		}

		if reply, handled := agents.HandlePendingWrite(r.Context(), req.UserID, req.SessionID, req.Message, dbConn); handled {
			writeJSON(w, chatResponse{SessionID: req.SessionID, Replies: []string{reply}})
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
		defer cancel()
		seq := selectedRunner.Run(ctx, req.UserID, req.SessionID, &genai.Content{
			Parts: []*genai.Part{{Text: req.Message}},
		}, agent.RunConfig{})

		var replies []string
		for event, err := range seq {
			if err != nil {
				log.Printf("chat error user=%s session=%s: %v", req.UserID, req.SessionID, err)
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			if event == nil || event.LLMResponse.Content == nil {
				continue
			}
			if event.Author == "user" {
				continue
			}
			for _, part := range event.LLMResponse.Content.Parts {
				if part == nil {
					continue
				}
				if part.Text != "" {
					replies = append(replies, part.Text)
				}
			}
		}

		if prompt, ok := agents.MaybeCaptureWrite(req.UserID, req.SessionID, replies); ok {
			writeJSON(w, chatResponse{SessionID: req.SessionID, Replies: []string{prompt}})
			return
		}

		writeJSON(w, chatResponse{SessionID: req.SessionID, Replies: replies})
	}
}

func chatDisabledHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusServiceUnavailable)
		writeJSON(w, chatResponse{
			Replies: []string{},
			Error:   "ai disabled: set ENABLE_AI=true to enable chat",
		})
	}
}

type metaResponse struct {
	AIEnabled bool `json:"ai_enabled"`
}

func metaHandler(aiEnabled bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		writeJSON(w, metaResponse{AIEnabled: aiEnabled})
	}
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.WriteHeader(status)
	writeJSON(w, errorResponse{Error: msg})
}

func dataHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
		snapshot, err := ckdb.GetSnapshot(ctx, dbConn)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to fetch snapshot")
			return
		}
		writeJSON(w, snapshot)
	}
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func normalizeAgentHint(msg string) string {
	trim := strings.TrimSpace(strings.ToLower(msg))
	// If user already sent an explicit hint, leave as-is.
	if strings.Contains(trim, "[agent:") || strings.Contains(trim, "option 1") || strings.Contains(trim, "option 2") || strings.Contains(trim, "option 3") || strings.Contains(trim, "option 4") {
		return msg
	}

	// Short direct responses.
	switch trim {
	case "1", "jobs", "job", "job applications", "job application":
		return "Option 1 (Jobs): " + msg
	case "2", "coding", "code", "leetcode", "problem":
		return "Option 2 (Coding): " + msg
	case "3", "projects", "project":
		return "Option 3 (Projects): " + msg
	case "4", "networking", "network":
		return "Option 4 (Networking): " + msg
	}

	// Keyword-based routing for full sentences.
	if strings.Contains(trim, "job") || strings.Contains(trim, "application") || strings.Contains(trim, "resume") {
		return "Option 1 (Jobs): " + msg
	}
	if strings.Contains(trim, "code") || strings.Contains(trim, "leetcode") || strings.Contains(trim, "problem") {
		return "Option 2 (Coding): " + msg
	}
	if strings.Contains(trim, "project") || strings.Contains(trim, "repo") || strings.Contains(trim, "portfolio") {
		return "Option 3 (Projects): " + msg
	}
	if strings.Contains(trim, "network") || strings.Contains(trim, "contact") || strings.Contains(trim, "coffee chat") {
		return "Option 4 (Networking): " + msg
	}

	return msg
}

func boolFromEnv(key string, defaultVal bool) bool {
	val := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if val == "" {
		return defaultVal
	}
	switch val {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	default:
		return defaultVal
	}
}

func resolveModelName(input, project, location string) (string, error) {
	name := strings.TrimSpace(input)
	if name == "" {
		name = "gemini-1.5-flash-002"
	}

	if strings.HasPrefix(name, "projects/") {
		if isValidVertexModelPath(name) {
			return name, nil
		}
		return "", fmt.Errorf("invalid Vertex model resource path: %q", name)
	}

	if strings.HasPrefix(name, "publishers/") {
		if isValidPublisherPath(name) {
			return fmt.Sprintf("projects/%s/locations/%s/%s", project, location, name), nil
		}
		return "", fmt.Errorf("invalid publisher model path: %q", name)
	}

	if suffix, ok := modelAliases[name]; ok {
		return fmt.Sprintf("projects/%s/locations/%s/%s", project, location, suffix), nil
	}

	return "", fmt.Errorf("unsupported model name %q; use a full Vertex resource path or one of: %s", name, strings.Join(sortedAliasKeys(), ", "))
}

func isValidVertexModelPath(path string) bool {
	return vertexModelPath.MatchString(path) || vertexPublisherPath.MatchString(path)
}

func isValidPublisherPath(path string) bool {
	return publisherPath.MatchString(path)
}

var (
	vertexModelPath        = regexp.MustCompile(`^projects/[^/]+/locations/[^/]+/models/[^/]+$`)
	vertexPublisherPath    = regexp.MustCompile(`^projects/[^/]+/locations/[^/]+/publishers/[^/]+/models/[^/]+$`)
	publisherPath          = regexp.MustCompile(`^publishers/[^/]+/models/[^/]+$`)
	vertexLocationFromPath = regexp.MustCompile(`^projects/[^/]+/locations/([^/]+)/`)
	modelAliases           = map[string]string{
		"gemini-1.5-flash-002": "publishers/google/models/gemini-1.5-flash-002",
		"gemini-1.5-pro-002":   "publishers/google/models/gemini-1.5-pro-002",
		"gemini-2.0-flash":     "publishers/google/models/gemini-2.0-flash",
		"gemini-2.0-pro":       "publishers/google/models/gemini-2.0-pro",
	}
)

func sortedAliasKeys() []string {
	keys := make([]string, 0, len(modelAliases))
	for k := range modelAliases {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func validateModelExists(ctx context.Context, modelName string) error {
	location := modelLocationFromPath(modelName)
	if location == "" {
		return fmt.Errorf("unable to detect location from model name: %q", modelName)
	}

	url := fmt.Sprintf("https://%s-aiplatform.googleapis.com/v1/%s", location, modelName)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	ts, err := google.DefaultTokenSource(ctx, "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return err
	}
	token, err := ts.Token()
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if resp.StatusCode == http.StatusOK {
		return nil
	}
	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("model not found in Vertex: %q", modelName)
	}
	msg := strings.TrimSpace(string(body))
	if msg == "" {
		msg = resp.Status
	}
	return fmt.Errorf("model validation failed: %s", msg)
}

func modelLocationFromPath(modelName string) string {
	if match := vertexLocationFromPath.FindStringSubmatch(modelName); match != nil {
		return match[1]
	}
	return ""
}

type jobCreateRequest struct {
	JobTitle   string `json:"job_title"`
	Company    string `json:"company"`
	JobLink    string `json:"job_link"`
	Status     string `json:"status"`
	Notes      string `json:"notes"`
	Applied    string `json:"applied_date"`
	ResultDate string `json:"result_date"`
}

type jobStatusUpdateRequest struct {
	ID     int64  `json:"id"`
	Status string `json:"status"`
}

func jobCreateHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var req jobCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}
		if strings.TrimSpace(req.JobTitle) == "" {
			writeError(w, http.StatusBadRequest, "job_title is required")
			return
		}
		var applied *time.Time
		if strings.TrimSpace(req.Applied) != "" {
			tm, err := time.Parse("2006-01-02", req.Applied)
			if err != nil {
				writeError(w, http.StatusBadRequest, "invalid applied_date")
				return
			}
			applied = &tm
		}
		var resultDate *time.Time
		if strings.TrimSpace(req.ResultDate) != "" {
			tm, err := time.Parse("2006-01-02", req.ResultDate)
			if err != nil {
				writeError(w, http.StatusBadRequest, "invalid result_date")
				return
			}
			resultDate = &tm
		}
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
		id, err := ckdb.InsertJobApplication(ctx, dbConn, ckdb.JobApplication{
			JobTitle:   req.JobTitle,
			Company:    req.Company,
			JobLink:    req.JobLink,
			Status:     req.Status,
			Notes:      req.Notes,
			Applied:    applied,
			ResultDate: resultDate,
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to create job")
			return
		}
		writeJSON(w, map[string]any{"id": id})
	}
}

func jobStatusUpdateHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var req jobStatusUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}
		if req.ID <= 0 {
			writeError(w, http.StatusBadRequest, "id is required")
			return
		}
		if strings.TrimSpace(req.Status) == "" {
			writeError(w, http.StatusBadRequest, "status is required")
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
		if err := ckdb.UpdateJobStatus(ctx, dbConn, req.ID, req.Status); err != nil {
			if err == sql.ErrNoRows {
				writeError(w, http.StatusNotFound, "job not found")
				return
			}
			if strings.Contains(err.Error(), "status is required") {
				writeError(w, http.StatusBadRequest, "status is required")
				return
			}
			writeError(w, http.StatusInternalServerError, "failed to update job status")
			return
		}
		writeJSON(w, map[string]any{"status": "updated"})
	}
}

type codingCreateRequest struct {
	LeetCodeNumber int    `json:"leetcode_number"`
	Title          string `json:"title"`
	Pattern        string `json:"pattern"`
	ProblemLink    string `json:"problem_link"`
	Difficulty     string `json:"difficulty"`
	AlreadySolved  bool   `json:"already_solved"`
	Notes          string `json:"notes"`
}

func codingCreateHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var req codingCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}
		if req.LeetCodeNumber == 0 && strings.TrimSpace(req.Title) == "" {
			writeError(w, http.StatusBadRequest, "leetcode_number or title is required")
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
		id, err := ckdb.InsertCodingProblem(ctx, dbConn, ckdb.CodingProblem{
			LeetCodeNumber: req.LeetCodeNumber,
			Title:          req.Title,
			Pattern:        req.Pattern,
			ProblemLink:    req.ProblemLink,
			Difficulty:     req.Difficulty,
			AlreadySolved:  req.AlreadySolved,
			Notes:          req.Notes,
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to create coding problem")
			return
		}
		writeJSON(w, map[string]any{"id": id})
	}
}

type projectCreateRequest struct {
	Name      string   `json:"name"`
	RepoURL   string   `json:"repo_url"`
	Active    bool     `json:"active"`
	TechStack []string `json:"tech_stack"`
	Summary   string   `json:"summary"`
}

func projectCreateHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var req projectCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}
		if strings.TrimSpace(req.Name) == "" {
			writeError(w, http.StatusBadRequest, "name is required")
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
		id, err := ckdb.InsertProject(ctx, dbConn, ckdb.Project{
			Name:      req.Name,
			RepoURL:   req.RepoURL,
			Active:    req.Active,
			TechStack: req.TechStack,
			Summary:   req.Summary,
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to create project")
			return
		}
		writeJSON(w, map[string]any{"id": id})
	}
}

type networkingCreateRequest struct {
	PersonName        string `json:"person_name"`
	HowMet            string `json:"how_met"`
	LinkedInConnected bool   `json:"linkedin_connected"`
	Company           string `json:"company"`
	Position          string `json:"position"`
	Notes             string `json:"notes"`
}

func networkingCreateHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var req networkingCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}
		if strings.TrimSpace(req.PersonName) == "" {
			writeError(w, http.StatusBadRequest, "person_name is required")
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
		id, err := ckdb.InsertNetworkingContact(ctx, dbConn, ckdb.NetworkingContact{
			PersonName:        req.PersonName,
			HowMet:            req.HowMet,
			LinkedInConnected: req.LinkedInConnected,
			Company:           req.Company,
			Position:          req.Position,
			Notes:             req.Notes,
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to create contact")
			return
		}
		writeJSON(w, map[string]any{"id": id})
	}
}

type goalUpdateRequest struct {
	Type        string `json:"type"`
	ID          int64  `json:"id"`
	Description string `json:"description"`
	Completed   bool   `json:"completed"`
}

func goalUpdateHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var req goalUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}
		if strings.TrimSpace(req.Type) == "" {
			writeError(w, http.StatusBadRequest, "type is required")
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
		if req.ID > 0 {
			if err := ckdb.UpdateGoal(ctx, dbConn, req.Type, req.ID, req.Completed, req.Description); err != nil {
				if err == sql.ErrNoRows {
					writeError(w, http.StatusNotFound, "goal not found")
					return
				}
				if strings.Contains(err.Error(), "invalid goal type") {
					writeError(w, http.StatusBadRequest, "invalid goal type")
					return
				}
				writeError(w, http.StatusInternalServerError, "failed to update goal")
				return
			}
			writeJSON(w, map[string]any{"status": "updated"})
			return
		}
		if strings.TrimSpace(req.Description) == "" {
			writeError(w, http.StatusBadRequest, "description is required")
			return
		}
		updated, err := ckdb.UpdateGoalCompletedByDescription(ctx, dbConn, req.Type, req.Description, req.Completed)
		if err != nil {
			if err == sql.ErrNoRows {
				writeError(w, http.StatusNotFound, "goal not found")
				return
			}
			if strings.Contains(err.Error(), "invalid goal type") {
				writeError(w, http.StatusBadRequest, "invalid goal type")
				return
			}
			writeError(w, http.StatusInternalServerError, "failed to update goal")
			return
		}
		writeJSON(w, map[string]any{"status": "updated", "updated": updated})
	}
}
