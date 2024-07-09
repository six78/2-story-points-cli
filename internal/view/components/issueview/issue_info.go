package issueview

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

const (
	loginPrefix = "@"
)

type issueInfo struct {
	err      error
	number   *int
	title    *string
	labels   []labelInfo
	author   *string
	assignee *string
}

type labelInfo struct {
	name  *string
	style lipgloss.Style
}

func renderNumber(info *issueInfo) string {
	if info == nil {
		return ""
	}
	if info.number == nil {
		return ""
	}
	return fmt.Sprintf("#%d", *info.number)
}

func authorString(info *issueInfo) string {
	if info == nil {
		return ""
	}

	if info.author == nil {
		return ""
	}

	return loginPrefix + *info.author
}

func assigneeString(info *issueInfo) string {
	if info == nil {
		return ""
	}

	if info.assignee == nil {
		return ""
	}

	return loginPrefix + *info.assignee
}
