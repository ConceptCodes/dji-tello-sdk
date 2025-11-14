package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// UniqueDirs removes duplicate and empty directory entries while preserving order.
func UniqueDirs(dirs ...string) []string {
	seen := make(map[string]struct{})
	var result []string
	for _, dir := range dirs {
		if dir == "" {
			continue
		}
		if _, exists := seen[dir]; exists {
			continue
		}
		seen[dir] = struct{}{}
		result = append(result, dir)
	}
	return result
}

// FindConfigFile searches through the provided directories for the first matching filename.
// Returns an empty string with nil error if no file exists.
func FindConfigFile(filenames []string, dirs []string) (string, error) {
	for _, dir := range dirs {
		for _, name := range filenames {
			candidate := filepath.Join(dir, name)
			info, err := os.Stat(candidate)
			if err == nil {
				if info.IsDir() {
					continue
				}
				return candidate, nil
			}
			if err != nil && !errors.Is(err, os.ErrNotExist) {
				return "", fmt.Errorf("failed to inspect config file %s: %w", candidate, err)
			}
		}
	}

	return "", nil
}
