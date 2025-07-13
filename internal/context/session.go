package context

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// SessionContext holds the current session state
type SessionContext struct {
	FocusedFiles    []string          `json:"focused_files"`
	OpenFiles       map[string]string `json:"open_files"`
	WorkingDir      string            `json:"working_dir"`
	ProjectRoot     string            `json:"project_root"`
	RecentFiles     []string          `json:"recent_files"`
	CurrentTask     string            `json:"current_task"`
	TaskHistory     []CompletedTask   `json:"task_history"`
	UserPreferences map[string]string `json:"user_preferences"`
	ProjectType     string            `json:"project_type"`

	SessionID       string            `json:"session_id"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
}

// CompletedTask represents a completed task in the session
type CompletedTask struct {
	Description string    `json:"description"`
	CompletedAt time.Time `json:"completed_at"`
	FilesChanged []string `json:"files_changed"`
}

// NewSessionContext creates a new session context
func NewSessionContext() *SessionContext {
	wd, _ := os.Getwd()
	return &SessionContext{
		FocusedFiles:    []string{},
		OpenFiles:       make(map[string]string),
		WorkingDir:      wd,
		ProjectRoot:     findProjectRoot(wd),
		RecentFiles:     []string{},
		TaskHistory:     []CompletedTask{},
		UserPreferences: make(map[string]string),

		SessionID:       generateSessionID(),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
}

// AddFocusedFile adds a file to the focused files list
func (sc *SessionContext) AddFocusedFile(filepath string) {
	// Remove if already exists
	sc.RemoveFocusedFile(filepath)
	// Add to beginning
	sc.FocusedFiles = append([]string{filepath}, sc.FocusedFiles...)
	// Keep only last 10 focused files
	if len(sc.FocusedFiles) > 10 {
		sc.FocusedFiles = sc.FocusedFiles[:10]
	}
	sc.UpdatedAt = time.Now()
}

// RemoveFocusedFile removes a file from focused files
func (sc *SessionContext) RemoveFocusedFile(filepath string) {
	for i, f := range sc.FocusedFiles {
		if f == filepath {
			sc.FocusedFiles = append(sc.FocusedFiles[:i], sc.FocusedFiles[i+1:]...)
			break
		}
	}
	sc.UpdatedAt = time.Now()
}

// ClearFocusedFiles clears all focused files
func (sc *SessionContext) ClearFocusedFiles() {
	sc.FocusedFiles = []string{}
	sc.UpdatedAt = time.Now()
}

// SetCurrentTask sets the current task
func (sc *SessionContext) SetCurrentTask(task string) {
	sc.CurrentTask = task
	sc.UpdatedAt = time.Now()
}

// CompleteCurrentTask marks the current task as completed
func (sc *SessionContext) CompleteCurrentTask(filesChanged []string) {
	if sc.CurrentTask != "" {
		completedTask := CompletedTask{
			Description:  sc.CurrentTask,
			CompletedAt:  time.Now(),
			FilesChanged: filesChanged,
		}
		sc.TaskHistory = append(sc.TaskHistory, completedTask)
		sc.CurrentTask = ""
		sc.UpdatedAt = time.Now()
	}
}

// AddRecentFile adds a file to recent files
func (sc *SessionContext) AddRecentFile(filepath string) {
	// Remove if already exists
	for i, f := range sc.RecentFiles {
		if f == filepath {
			sc.RecentFiles = append(sc.RecentFiles[:i], sc.RecentFiles[i+1:]...)
			break
		}
	}
	// Add to beginning
	sc.RecentFiles = append([]string{filepath}, sc.RecentFiles...)
	// Keep only last 20 recent files
	if len(sc.RecentFiles) > 20 {
		sc.RecentFiles = sc.RecentFiles[:20]
	}
	sc.UpdatedAt = time.Now()
}



// GetContextSummary returns a summary of the current context
func (sc *SessionContext) GetContextSummary() string {
	var summary strings.Builder
	
	summary.WriteString(fmt.Sprintf("Session Context Summary\n"))
	summary.WriteString(fmt.Sprintf("======================\n"))
	summary.WriteString(fmt.Sprintf("Working Directory: %s\n", sc.WorkingDir))
	
	if sc.ProjectRoot != "" {
		summary.WriteString(fmt.Sprintf("Project Root: %s\n", sc.ProjectRoot))
	}
	
	if sc.ProjectType != "" {
		summary.WriteString(fmt.Sprintf("Project Type: %s\n", sc.ProjectType))
	}
	
	if sc.CurrentTask != "" {
		summary.WriteString(fmt.Sprintf("Current Task: %s\n", sc.CurrentTask))
	}
	
	if len(sc.FocusedFiles) > 0 {
		summary.WriteString(fmt.Sprintf("Focused Files (%d):\n", len(sc.FocusedFiles)))
		for i, file := range sc.FocusedFiles {
			if i < 5 { // Show only first 5
				summary.WriteString(fmt.Sprintf("  - %s\n", file))
			}
		}
		if len(sc.FocusedFiles) > 5 {
			summary.WriteString(fmt.Sprintf("  ... and %d more\n", len(sc.FocusedFiles)-5))
		}
	}
	
	if len(sc.RecentFiles) > 0 {
		summary.WriteString(fmt.Sprintf("Recent Files (%d):\n", len(sc.RecentFiles)))
		for i, file := range sc.RecentFiles {
			if i < 3 { // Show only first 3
				summary.WriteString(fmt.Sprintf("  - %s\n", file))
			}
		}
		if len(sc.RecentFiles) > 3 {
			summary.WriteString(fmt.Sprintf("  ... and %d more\n", len(sc.RecentFiles)-3))
		}
	}
	
	if len(sc.TaskHistory) > 0 {
		summary.WriteString(fmt.Sprintf("Completed Tasks: %d\n", len(sc.TaskHistory)))
		// Show last completed task
		lastTask := sc.TaskHistory[len(sc.TaskHistory)-1]
		summary.WriteString(fmt.Sprintf("  Last: %s (completed %s)\n", 
			lastTask.Description, 
			lastTask.CompletedAt.Format("15:04")))
	}
	

	
	return summary.String()
}

//  saves the session context to a file
func (sc *SessionContext) SaveToFile(filename string) error {
	data, err := json.MarshalIndent(sc, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}

// loads session context from a file
func LoadFromFile(filename string) (*SessionContext, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	
	var sc SessionContext
	err = json.Unmarshal(data, &sc)
	if err != nil {
		return nil, err
	}
	
	return &sc, nil
}

// Helpers

func findProjectRoot(startDir string) string {
	dir := startDir
	for {
		// Check for common project root indicators
		indicators := []string{".git", "go.mod", "package.json", "Cargo.toml", "pyproject.toml", "requirements.txt"}
		for _, indicator := range indicators {
			if _, err := os.Stat(filepath.Join(dir, indicator)); err == nil {
				return dir
			}
		}
		
		parent := filepath.Dir(dir)
		if parent == dir {
			break // Reached root
		}
		dir = parent
	}
	return startDir // Fallback to start directory
}

func generateSessionID() string {
	return fmt.Sprintf("session_%d", time.Now().Unix())
}

// DetectProjectType detects the project type based on files in the project root
func (sc *SessionContext) DetectProjectType() {
	if sc.ProjectRoot == "" {
		return
	}
	
	// Check for Go project
	if _, err := os.Stat(filepath.Join(sc.ProjectRoot, "go.mod")); err == nil {
		sc.ProjectType = "go"
		return
	}
	
	// Check for Node.js project
	if _, err := os.Stat(filepath.Join(sc.ProjectRoot, "package.json")); err == nil {
		sc.ProjectType = "nodejs"
		return
	}
	
	// Check for Python project
	indicators := []string{"requirements.txt", "pyproject.toml", "setup.py"}
	for _, indicator := range indicators {
		if _, err := os.Stat(filepath.Join(sc.ProjectRoot, indicator)); err == nil {
			sc.ProjectType = "python"
			return
		}
	}
	
	// Check for Rust project
	if _, err := os.Stat(filepath.Join(sc.ProjectRoot, "Cargo.toml")); err == nil {
		sc.ProjectType = "rust"
		return
	}
	
	sc.ProjectType = "unknown"
}
