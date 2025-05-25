package main

import (
	"os"
	"testing"
)

func TestOpenDatabase_NoEnv(t *testing.T) {
	old := os.Getenv("DATABASE_URL")
	os.Unsetenv("DATABASE_URL")
	defer os.Setenv("DATABASE_URL", old)

	if _, err := OpenDatabase(); err == nil {
		t.Fatal("expected error when DATABASE_URL is empty")
	}
}
