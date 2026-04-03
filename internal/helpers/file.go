package helpers

import (
	"fmt"
	"io"
	"os"
)

// CopyFile copies a single file from src to dst.
func CopyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open %s: %w", src, err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create %s: %w", dst, err)
	}

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		dstFile.Close()
		return fmt.Errorf("copy to %s: %w", dst, err)
	}

	if err := dstFile.Close(); err != nil {
		return fmt.Errorf("close %s: %w", dst, err)
	}

	return nil
}
