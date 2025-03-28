// Package tests include helper code for tests
package tests

import (
	"os"
	"path/filepath"
	"runtime"
)

func LoadData(name string) string {
	_, f, _, _ := runtime.Caller(0)
	fp := filepath.Join(filepath.Dir(f), "data", name)

	r, err := os.ReadFile(fp)

	if err != nil {
		panic(err)
	}

	return string(r)
}
