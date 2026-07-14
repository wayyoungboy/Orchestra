package security

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrCommandNotAllowed = errors.New("command not in whitelist")
	ErrPathNotAllowed    = errors.New("path not in whitelist")
)

type Whitelist struct {
	commands map[string]bool
	paths    []string
}

func NewWhitelist(commands, paths []string) *Whitelist {
	cmdMap := make(map[string]bool)
	for _, cmd := range commands {
		cmd = strings.TrimSpace(cmd)
		if cmd == "" {
			continue
		}
		if filepath.IsAbs(cmd) {
			if canonical, err := canonicalPath(cmd); err == nil {
				cmdMap[canonical] = true
			}
			continue
		}
		cmdMap[cmd] = true
	}

	expandedPaths := make([]string, 0, len(paths))
	for _, path := range paths {
		expandedPaths = append(expandedPaths, expandPath(path))
	}

	return &Whitelist{
		commands: cmdMap,
		paths:    expandedPaths,
	}
}

func expandPath(path string) string {
	if path != "~" && !strings.HasPrefix(path, "~"+string(filepath.Separator)) {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	if path == "~" {
		return home
	}
	return filepath.Join(home, path[2:])
}

func (w *Whitelist) ValidateCommand(cmd string) error {
	if filepath.IsAbs(cmd) {
		canonical, err := canonicalPath(cmd)
		if err != nil || !w.commands[canonical] {
			return ErrCommandNotAllowed
		}
		return nil
	}

	// A command name may be approved for PATH lookup, but a relative path such
	// as ./codex or ../codex must not inherit that approval. Otherwise a member
	// could replace a binary with an arbitrary executable having the same base
	// name.
	if filepath.Base(cmd) != cmd || !w.commands[cmd] {
		return ErrCommandNotAllowed
	}
	return nil
}

func (w *Whitelist) ValidatePath(path string) error {
	absPath, err := canonicalPath(path)
	if err != nil {
		return err
	}

	for _, allowed := range w.paths {
		absAllowed, err := canonicalPath(allowed)
		if err != nil {
			continue
		}
		rel, err := filepath.Rel(absAllowed, absPath)
		if err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel) {
			return nil
		}
	}

	return ErrPathNotAllowed
}

func canonicalPath(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	absPath = filepath.Clean(absPath)

	// Resolve the deepest existing parent. This prevents a lexical path check
	// from being bypassed by a symlink inside an allowed directory, even when
	// the final path has not been created yet.
	for current := absPath; ; current = filepath.Dir(current) {
		if resolved, err := filepath.EvalSymlinks(current); err == nil {
			rel, err := filepath.Rel(current, absPath)
			if err != nil {
				return "", err
			}
			return filepath.Join(resolved, rel), nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			return absPath, nil
		}
	}
}

func (w *Whitelist) AllowedCommands() []string {
	result := make([]string, 0, len(w.commands))
	for cmd := range w.commands {
		result = append(result, cmd)
	}
	return result
}

func (w *Whitelist) AllowedPaths() []string {
	return w.paths
}
