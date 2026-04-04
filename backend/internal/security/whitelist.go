package security

import (
	"errors"
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
		cmdMap[filepath.Base(cmd)] = true
	}

	return &Whitelist{
		commands: cmdMap,
		paths:    paths,
	}
}

func (w *Whitelist) ValidateCommand(cmd string) error {
	base := filepath.Base(cmd)
	if !w.commands[base] {
		return ErrCommandNotAllowed
	}
	return nil
}

func (w *Whitelist) ValidatePath(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	for _, allowed := range w.paths {
		absAllowed, err := filepath.Abs(allowed)
		if err != nil {
			continue
		}
		if strings.HasPrefix(absPath, absAllowed) {
			return nil
		}
	}

	return ErrPathNotAllowed
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