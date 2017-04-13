// +build !windows

package client

import (
	"path/filepath"
)

func getContextRoot(srcPath string) (string, error) {
	return filepath.Join(srcPath, "."), nil
}

func convertToUnixPath(path string) (int, string) {
	return 0, path
}

func recoverPath(pathType int, path string) string {
	return path
}
