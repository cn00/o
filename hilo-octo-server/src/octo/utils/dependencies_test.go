package utils

import (
	"reflect"
	"testing"
)

func TestSplitDependencies(t *testing.T) {
	patterns := []struct {
		s        string
		expected []int
	}{
		{
			s:        "",
			expected: nil,
		},
		{
			s:        "1",
			expected: []int{1},
		},
		{
			s:        "1,2",
			expected: []int{1, 2},
		},
		{
			s:        ",1,,",
			expected: []int{1},
		},
	}

	for i, pattern := range patterns {
		deps, err := SplitDependencies(pattern.s)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(deps, pattern.expected) {
			t.Fatalf("#%d: SplitDependencies is wrong: %v", i, deps)
		}
	}
}

func TestJoinDependencies(t *testing.T) {
	s := JoinDependencies([]int{1, 2})
	if s != "1,2" {
		t.Fatal("JoinDependencies is wrong:", s)
	}
}
