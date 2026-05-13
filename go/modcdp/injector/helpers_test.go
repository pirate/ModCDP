package injector_test

import (
	"io"
	"os"
	"path/filepath"
)

func boolPtr(value bool) *bool {
	return &value
}

func copyDir(src string, dst string) error {
	return filepath.WalkDir(src, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(dst, relPath)
		if entry.IsDir() {
			return os.MkdirAll(targetPath, 0o755)
		}
		source, err := os.Open(path)
		if err != nil {
			return err
		}
		defer source.Close()
		info, err := entry.Info()
		if err != nil {
			return err
		}
		target, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode())
		if err != nil {
			return err
		}
		defer target.Close()
		_, err = io.Copy(target, source)
		return err
	})
}
