package speckit

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Spec represents parsed spec.md content
type Spec struct {
	Title        string
	Description  string
	UserStories  []UserStory
	Requirements []string
	RawContent   string
	FilePath     string
}

type UserStory struct {
	ID          string
	Title       string
	Description string
	Acceptance  []string
}

// Plan represents parsed plan.md content
type Plan struct {
	Overview     string
	TechStack    []string
	Architecture string
	Phases       []PlanPhase
	RawContent   string
	FilePath     string
}

type PlanPhase struct {
	Name        string
	Description string
	Steps       []string
}

// TaskItem represents a single task from tasks.md
type TaskItem struct {
	ID           string
	Title        string
	Description  string
	UserStoryRef string
	Dependencies []string
	FilePaths    []string
	IsParallel   bool
	IsTest       bool
	Order        int
}

// Question represents a clarification question
type Question struct {
	ID       string
	Question string
	Context  string
	Answered bool
	Answer   string
}

// ParseSpec reads and parses a spec.md file
func ParseSpec(featureDir string) (*Spec, error) {
	specPath := filepath.Join(featureDir, "spec.md")

	content, err := os.ReadFile(specPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read spec.md: %w", err)
	}

	spec := &Spec{
		RawContent: string(content),
		FilePath:   specPath,
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "# ") {
			spec.Title = strings.TrimPrefix(line, "# ")
			break
		}
	}

	userStoryRegex := regexp.MustCompile(`(?m)^##\s*User Story:?\s*(.+)$`)
	matches := userStoryRegex.FindAllStringSubmatch(string(content), -1)

	for i, match := range matches {
		if len(match) >= 2 {
			spec.UserStories = append(spec.UserStories, UserStory{
				ID:    fmt.Sprintf("US-%d", i+1),
				Title: strings.TrimSpace(match[1]),
			})
		}
	}

	return spec, nil
}

// ParsePlan reads and parses a plan.md file
func ParsePlan(featureDir string) (*Plan, error) {
	planPath := filepath.Join(featureDir, "plan.md")

	content, err := os.ReadFile(planPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read plan.md: %w", err)
	}

	plan := &Plan{
		RawContent: string(content),
		FilePath:   planPath,
	}

	inTechStack := false
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		lower := strings.ToLower(line)
		if strings.Contains(lower, "tech stack") || strings.Contains(lower, "technology") {
			inTechStack = true
			continue
		}
		if inTechStack && strings.HasPrefix(line, "##") {
			inTechStack = false
		}
		if inTechStack && strings.HasPrefix(line, "- ") {
			plan.TechStack = append(plan.TechStack, strings.TrimPrefix(line, "- "))
		}
	}

	return plan, nil
}

// ParseTasks reads and parses a tasks.md file
func ParseTasks(featureDir string) ([]TaskItem, error) {
	tasksPath := filepath.Join(featureDir, "tasks.md")

	file, err := os.Open(tasksPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open tasks.md: %w", err)
	}
	defer file.Close()

	var tasks []TaskItem
	var currentUserStory string
	taskOrder := 0

	scanner := bufio.NewScanner(file)
	taskRegex := regexp.MustCompile(`^[-*]\s+\[[ xX]?\]\s*(.+)$`)
	parallelRegex := regexp.MustCompile(`\[P\]`)
	userStoryRegex := regexp.MustCompile(`(?i)^##\s*(User Story|Phase|Story):?\s*(.+)$`)
	filePathRegex := regexp.MustCompile("`([^`]+\\.[a-zA-Z]+)`")

	for scanner.Scan() {
		line := scanner.Text()

		if matches := userStoryRegex.FindStringSubmatch(line); len(matches) >= 3 {
			currentUserStory = strings.TrimSpace(matches[2])
			continue
		}

		if matches := taskRegex.FindStringSubmatch(line); len(matches) >= 2 {
			taskOrder++
			taskTitle := strings.TrimSpace(matches[1])

			isParallel := parallelRegex.MatchString(taskTitle)
			taskTitle = parallelRegex.ReplaceAllString(taskTitle, "")
			taskTitle = strings.TrimSpace(taskTitle)

			isTest := strings.Contains(strings.ToLower(taskTitle), "test")

			var filePaths []string
			if fpMatches := filePathRegex.FindAllStringSubmatch(taskTitle, -1); len(fpMatches) > 0 {
				for _, fp := range fpMatches {
					if len(fp) >= 2 {
						filePaths = append(filePaths, fp[1])
					}
				}
			}

			task := TaskItem{
				ID:           fmt.Sprintf("T-%03d", taskOrder),
				Title:        taskTitle,
				UserStoryRef: currentUserStory,
				FilePaths:    filePaths,
				IsParallel:   isParallel,
				IsTest:       isTest,
				Order:        taskOrder,
			}

			tasks = append(tasks, task)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning tasks.md: %w", err)
	}

	return tasks, nil
}

// ParseClarifications extracts questions from clarify output
func ParseClarifications(output string) []Question {
	var questions []Question

	questionRegex := regexp.MustCompile(`(?m)^\d+\.\s+(.+\?)`)
	matches := questionRegex.FindAllStringSubmatch(output, -1)

	for i, match := range matches {
		if len(match) >= 2 {
			questions = append(questions, Question{
				ID:       fmt.Sprintf("Q%d", i+1),
				Question: strings.TrimSpace(match[1]),
			})
		}
	}

	return questions
}

// Summary returns a brief summary for Telegram
func (s *Spec) Summary() string {
	summary := fmt.Sprintf("*%s*\n\n", s.Title)

	if len(s.UserStories) > 0 {
		summary += fmt.Sprintf("User Stories: %d\n", len(s.UserStories))
		for _, us := range s.UserStories {
			title := us.Title
			if len(title) > 50 {
				title = title[:47] + "..."
			}
			summary += fmt.Sprintf("  - %s\n", title)
		}
	}

	return summary
}

// Summary returns a brief summary for Telegram
func (p *Plan) Summary() string {
	summary := "*Implementation Plan*\n\n"

	if len(p.TechStack) > 0 {
		summary += "Tech Stack:\n"
		for _, tech := range p.TechStack {
			summary += fmt.Sprintf("  - %s\n", tech)
		}
	}

	if len(p.Phases) > 0 {
		summary += fmt.Sprintf("\nPhases: %d\n", len(p.Phases))
	}

	return summary
}
