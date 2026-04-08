package filesystem

import (
	"os"
	"path/filepath"
	"sort"
	"time"
)

type FileInfo struct {
	Name    string    `json:"name"`
	Path    string    `json:"path"`
	IsDir   bool      `json:"isDir"`
	Size    int64     `json:"size"`
	ModTime time.Time `json:"modTime"`
	Mode    string    `json:"mode"`
}

type Browser struct {
	validator *Validator
}

func NewBrowser(validator *Validator) *Browser {
	return &Browser{validator: validator}
}

// ListDir lists entries under path. If dirsOnly is true, regular files are omitted (for picking a working directory).
func (b *Browser) ListDir(path string, dirsOnly bool) ([]*FileInfo, error) {
	if err := b.validator.ValidatePath(path); err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var files []*FileInfo
	for _, entry := range entries {
		full := filepath.Join(path, entry.Name())
		if dirsOnly {
			fi, err := os.Stat(full)
			if err != nil || !fi.IsDir() {
				continue
			}
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}

		isDir := entry.IsDir()
		if dirsOnly {
			isDir = true
		}

		files = append(files, &FileInfo{
			Name:    entry.Name(),
			Path:    full,
			IsDir:   isDir,
			Size:    info.Size(),
			ModTime: info.ModTime(),
			Mode:    info.Mode().String(),
		})
	}

	sortFiles(files)
	return files, nil
}

func (b *Browser) GetHomeDir() (string, error) {
	return os.UserHomeDir()
}

func (b *Browser) PathExists(path string) (bool, error) {
	if err := b.validator.ValidatePath(path); err != nil {
		return false, err
	}
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	return err == nil, err
}

func sortFiles(files []*FileInfo) {
	sort.Slice(files, func(i, j int) bool {
		if files[i].IsDir != files[j].IsDir {
			return files[i].IsDir
		}
		return files[i].Name < files[j].Name
	})
}

// PathValidationResult contains detailed path validation results
type PathValidationResult struct {
	Exists    bool   `json:"exists"`
	Readable  bool   `json:"readable"`
	Writable  bool   `json:"writable"`
	IsDir     bool   `json:"isDir"`
	Error     string `json:"error,omitempty"`
	Validated bool   `json:"validated"`
}

// ValidatePath performs comprehensive path validation
func (b *Browser) ValidatePath(path string) *PathValidationResult {
	result := &PathValidationResult{Validated: true}

	// Check if path is allowed by validator
	if err := b.validator.ValidatePath(path); err != nil {
		result.Error = err.Error()
		return result
	}

	// Check if path exists
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		result.Error = "path does not exist"
		return result
	}
	if err != nil {
		result.Error = err.Error()
		return result
	}

	result.Exists = true
	result.IsDir = info.IsDir()

	// Check readability
	file, err := os.Open(path)
	if err == nil {
		result.Readable = true
		file.Close()
	}

	// Check writability
	if result.IsDir {
		// For directories, try to create a temp file
		testFile := filepath.Join(path, ".orchestra_write_test")
		f, err := os.Create(testFile)
		if err == nil {
			result.Writable = true
			f.Close()
			os.Remove(testFile)
		}
	} else {
		// For files, check if we can open for appending
		f, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND, 0)
		if err == nil {
			result.Writable = true
			f.Close()
		}
	}

	return result
}