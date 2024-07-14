package votestate

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/six78/2-story-points-cli/internal/config"
	"github.com/six78/2-story-points-cli/internal/view/messages"
)

var style = lipgloss.NewStyle().Foreground(config.ForegroundShadeColor)

type Model struct {
	duration time.Duration
	start    time.Time
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case messages.AutoRevealScheduled:
		m.start = time.Now()
		m.duration = msg.Duration
	case messages.AutoRevealCancelled:
		m.start = time.Time{}
		m.duration = 0
	}

	return m, nil
}

func (m Model) View() string {
	if m.start.IsZero() {
		return ""
	}
	left := (m.duration - time.Since(m.start)).Seconds()
	if left == 0 {
		return style.Render("Revealing votes...")
	}
	return style.Render(fmt.Sprintf("Revealing in %.1f", left))
}
