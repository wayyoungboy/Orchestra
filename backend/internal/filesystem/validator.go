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
	absPath, err := canonicalPath(path)
	if err != nil {
		return err
	}

	for _, allowed := range v.allowedPaths {
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
	// from being bypassed by a symlink inside an allowed directory.
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
