package utils

import (
	"reflect"
	"testing"
)

func TestSplitTags(t *testing.T) {
	patterns := []struct {
		s        string
		expected []string
	}{
		{
			s:        "",
			expected: []string{},
		},
		{
			s:        "tag",
			expected: []string{"tag"},
		},
		{
			s:        "tag1,tag2",
			expected: []string{"tag1", "tag2"},
		},
		{
			s:        ",tag,,",
			expected: []string{"tag"},
		},
	}

	for i, pattern := range patterns {
		tags := SplitTags(pattern.s)
		if !reflect.DeepEqual(tags, pattern.expected) {
			t.Fatalf("#%d: SplitTags is wrong: %v", i, tags)
		}
	}
}

func TestJoinTags(t *testing.T) {
	s := JoinTags([]string{"tag1", "tag2"})
	if s != "tag1,tag2" {
		t.Fatal("JoinTags is wrong:", s)
	}
}

func TestMergeTags(t *testing.T) {
	patterns := []struct {
		tags     []string
		addTags  []string
		expected []string
	}{
		{
			tags:     nil,
			addTags:  nil,
			expected: nil,
		},
		{
			tags:     []string{"tag1"},
			addTags:  []string{"tag2"},
			expected: []string{"tag1", "tag2"},
		},
		{
			tags:     nil,
			addTags:  []string{"tag1"},
			expected: []string{"tag1"},
		},
		{
			tags:     []string{"tag1"},
			addTags:  nil,
			expected: []string{"tag1"},
		},
		{
			tags:     []string{"tag1"},
			addTags:  []string{"tag1"},
			expected: []string{"tag1"},
		},
	}

	for i, pattern := range patterns {
		newTags := MergeTags(pattern.tags, pattern.addTags)
		if !reflect.DeepEqual(newTags, pattern.expected) {
			t.Fatalf("#%d: newTags is wrong: %v", i, newTags)
		}
	}
}

func TestRemoveTags(t *testing.T) {
	patterns := []struct {
		tags       []string
		removeTags []string
		expected   []string
	}{
		{
			tags:       nil,
			removeTags: nil,
			expected:   nil,
		},
		{
			tags:       []string{"tag1", "tag2"},
			removeTags: []string{"tag2"},
			expected:   []string{"tag1"},
		},
		{
			tags:       []string{"tag1"},
			removeTags: []string{"tag1"},
			expected:   []string{},
		},
		{
			tags:       []string{"tag1"},
			removeTags: nil,
			expected:   []string{"tag1"},
		},
		{
			tags:       []string{"tag1"},
			removeTags: []string{"tag2"},
			expected:   []string{"tag1"},
		},
	}

	for i, pattern := range patterns {
		newTags := RemoveTags(pattern.tags, pattern.removeTags)
		if !reflect.DeepEqual(newTags, pattern.expected) {
			t.Fatalf("#%d: newTags is wrong: %v", i, newTags)
		}
	}
}
