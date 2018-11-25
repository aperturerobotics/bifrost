package main

import (
	"testing"
)

func TestFileTransfer(t *testing.T) {
	if err := doIt(false); err != nil {
		t.Fatal(err.Error())
	}
}
