package cli

import (
	"os"
	"path/filepath"
)

// getOrchestraSkillsDir returns the path to Orchestra's local skills directory (~/.orchestra/skills).
func getOrchestraSkillsDir() string {
	home, _ := os.UserHomeDir()
	if home == "" {
		return "~/.orchestra/skills"
	}
	return filepath.Join(home, ".orchestra", "skills")
}
