package main

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestStringEqualityWithLastIndexFunctionOnPaths(t *testing.T) {
	tests := []struct {
		name string
		in   string
		out  string
	}{
		{
			name: "Should equal to",
			in:   "/var/run/secrets.txt",
			out:  "/var/run",
		},
	}

	for _, v := range tests {
		t.Run(v.name, func(t *testing.T) {
			lastReverseSlashIndex := strings.LastIndex(v.in, "/")

			folderPath := v.in[:lastReverseSlashIndex]

			assert.Equal(t, v.out, folderPath)
		})
	}
}
