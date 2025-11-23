package utils

import "os"

func CleanupFiles(dir string) {
	_ = os.RemoveAll(dir)
}
