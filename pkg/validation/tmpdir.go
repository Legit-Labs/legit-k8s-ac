package validation

import (
	"os"
)

func withTmpDir(foo func(tmpDir string) error) error {
	dir, err := os.MkdirTemp("", "admission-controller-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	err = foo(dir)
	if err != nil {
		return err
	}

	return nil
}
