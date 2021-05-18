package service

import "testing"

func TestMatchFilename(t *testing.T) {
	patterns := []struct {
		filename string
		names    []string
		expected bool
	}{
		{
			filename: "file1",
			names:    []string{"file", "file1"},
			expected: true,
		},
		{
			filename: "file1",
			names:    []string{"file", "file2"},
			expected: false,
		},
		{
			filename: "file1",
			names:    nil,
			expected: true,
		},
	}

	for i, pattern := range patterns {
		matched := matchFilename(pattern.filename, pattern.names)
		if matched != pattern.expected {
			t.Fatalf("#%d: matchFilename is wrong: %v", i, matched)
		}
	}
}
