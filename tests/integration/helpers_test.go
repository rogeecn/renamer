package integration

import "os"

func fileExistsTestHelper(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
