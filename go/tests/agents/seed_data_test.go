package agenttests

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

type seedJob struct {
	Title   string
	Company string
}

type seedCoding struct {
	Title   string
	Pattern string
}

type seedProject struct {
	Name      string
	TechStack []string
}

type seedContact struct {
	Name    string
	Company string
}

type seedData struct {
	Jobs     []seedJob
	Coding   []seedCoding
	Projects []seedProject
	Contacts []seedContact
}

var (
	seedOnce  sync.Once
	seedCache seedData
	seedErr   error
)

func loadSeedData(t *testing.T) seedData {
	t.Helper()
	seedOnce.Do(func() {
		seedCache, seedErr = loadSeedDataFromFile()
	})
	if seedErr != nil {
		t.Fatalf("load seed data: %v", seedErr)
	}
	return seedCache
}

func loadSeedDataFromFile() (seedData, error) {
	path := filepath.Join("..", "..", "migrations", "20251201_000001_seed_sample.sql")
	file, err := os.Open(path)
	if err != nil {
		return seedData{}, err
	}
	defer file.Close()

	var data seedData
	var mode string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "--") {
			continue
		}
		switch {
		case strings.HasPrefix(line, "INSERT INTO job_applications"):
			mode = "jobs"
			continue
		case strings.HasPrefix(line, "INSERT INTO coding_problems"):
			mode = "coding"
			continue
		case strings.HasPrefix(line, "INSERT INTO projects"):
			mode = "projects"
			continue
		case strings.HasPrefix(line, "INSERT INTO networking_contacts"):
			mode = "contacts"
			continue
		case strings.HasPrefix(line, "INSERT INTO"):
			mode = ""
			continue
		}

		if mode == "" || !strings.HasPrefix(line, "(") {
			continue
		}

		values := parseQuotedValues(line)
		switch mode {
		case "jobs":
			if len(values) >= 2 {
				data.Jobs = append(data.Jobs, seedJob{
					Title:   values[0],
					Company: values[1],
				})
			}
		case "coding":
			if len(values) >= 2 {
				data.Coding = append(data.Coding, seedCoding{
					Title:   values[0],
					Pattern: values[1],
				})
			}
		case "projects":
			if len(values) >= 3 {
				data.Projects = append(data.Projects, seedProject{
					Name:      values[0],
					TechStack: values[2 : len(values)-1],
				})
			}
		case "contacts":
			if len(values) >= 3 {
				data.Contacts = append(data.Contacts, seedContact{
					Name:    values[0],
					Company: values[2],
				})
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return seedData{}, err
	}

	if len(data.Jobs) == 0 || len(data.Coding) == 0 || len(data.Projects) == 0 || len(data.Contacts) == 0 {
		return seedData{}, errSeedDataMissing(data)
	}

	return data, nil
}

func errSeedDataMissing(data seedData) error {
	return fmt.Errorf("seed data missing: jobs=%d coding=%d projects=%d contacts=%d", len(data.Jobs), len(data.Coding), len(data.Projects), len(data.Contacts))
}

func parseQuotedValues(line string) []string {
	parts := strings.Split(line, "'")
	values := make([]string, 0, len(parts)/2)
	for i := 1; i < len(parts); i += 2 {
		values = append(values, parts[i])
	}
	return values
}
