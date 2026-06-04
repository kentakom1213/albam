package api

import (
	"fmt"
	"path/filepath"
	"strings"
)

func ResolveUnderRoot(root string, relPath string) (string, error) {
	if root == "" {
		return "", fmt.Errorf("root is empty")
	}
	if relPath == "" {
		return "", fmt.Errorf("relative path is empty")
	}
	if filepath.IsAbs(relPath) {
		return "", fmt.Errorf("absolute path is not allowed: %s", relPath)
	}

	cleanRel := filepath.Clean(relPath)
	if cleanRel == "." || cleanRel == ".." || strings.HasPrefix(cleanRel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path escapes root: %s", relPath)
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("resolve root: %w", err)
	}

	target := filepath.Join(absRoot, cleanRel)
	absTarget, err := filepath.Abs(target)
	if err != nil {
		return "", fmt.Errorf("resolve target: %w", err)
	}

	rel, err := filepath.Rel(absRoot, absTarget)
	if err != nil {
		return "", fmt.Errorf("check relative path: %w", err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return "", fmt.Errorf("path escapes root: %s", relPath)
	}

	return absTarget, nil
}

func SafeCacheID(id string) bool {
	if id == "" {
		return false
	}

	for _, r := range id {
		switch {
		case 'a' <= r && r <= 'z':
		case 'A' <= r && r <= 'Z':
		case '0' <= r && r <= '9':
		case r == '-':
		case r == '_':
		default:
			return false
		}
	}

	return true
}
