package filesystem

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrPathNotAllowed = errors.New("path not allowed")
	ErrPathNotExist   = errors.New("path does not exist")
)

type Validator struct {
	allowedPaths []string
}

func NewValidator(allowedPaths []string) *Validator {
	expanded := make([]string, 0, len(allowedPaths))
	for _, p := range allowedPaths {
		expanded = append(expanded, expandPath(p))
	}
	return &Validator{allowedPaths: expanded}
}

func (v *Validator) ValidatePath(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	for _, allowed := range v.allowedPaths {
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

func (v *Validator) ValidateExists(path string) error {
	if err := v.ValidatePath(path); err != nil {
		return err
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return ErrPathNotExist
	}
	return nil
}

func expandPath(path string) string {
	if len(path) > 0 && path[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[1:])
	}
	return path
}