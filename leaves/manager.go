package leaves

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Manager struct {
	leaveDays map[string]bool
}

func NewManager() *Manager {
	m := &Manager{
		leaveDays: make(map[string]bool),
	}
	m.loadLeaves()
	return m
}

func (m *Manager) loadLeaves() {
	// Try to load leaves from a file in the project root or home directory
	leavesFile := "leaves.txt"

	// Check if local leaves.txt exists
	if _, err := os.Stat(leavesFile); os.IsNotExist(err) {
		// Try home directory
		home, _ := os.UserHomeDir()
		leavesFile = filepath.Join(home, ".jiralogs", "leaves.txt")
	}

	file, err := os.Open(leavesFile)
	if err != nil {
		// No leaves file, that's OK
		return
	}
	defer func() {
		_ = file.Close()
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse date in format YYYY-MM-DD
		if t, err := time.Parse("2006-01-02", line); err == nil {
			m.leaveDays[t.Format("2006-01-02")] = true
		}
	}
}

func (m *Manager) IsLeaveDay(t time.Time) bool {
	return m.leaveDays[t.Format("2006-01-02")]
}
