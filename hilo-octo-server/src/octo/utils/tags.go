package utils

import (
	"strings"
)

func SplitTags(s string) []string {
	return strings.FieldsFunc(s, isComma)
}

func isComma(r rune) bool {
	return r == ','
}

func JoinTags(tags []string) string {
	return strings.Join(tags, ",")
}

func MergeTags(tags, addTags []string) []string {
	newTags := tags
	for _, addTag := range addTags {
		contain := false
		for _, tag := range tags {
			if tag == addTag {
				contain = true
				break
			}
		}
		if !contain {
			newTags = append(newTags, addTag)
		}
	}
	return newTags
}

func RemoveTags(tags []string, removeTags []string) []string {

	if tags == nil {
		return nil
	}

	result := []string{}
	for _, tag := range tags {
		contains := false
		for _, removeTag := range removeTags {
			if tag == removeTag {
				contains = true
				break
			}
		}
		if !contains {
			result = append(result, tag)
		}
	}

	return result
}
