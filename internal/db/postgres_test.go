package db

import "testing"

func TestNewDB(t *testing.T) {
	_, err := New("postgres://user:password@127.0.0.1:5432/spydb?sslmode=disable")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
