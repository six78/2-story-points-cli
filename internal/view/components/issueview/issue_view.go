package issueview

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/go-github/v61/github"
	"go.uber.org/zap"

	"github.com/six78/2-story-points-cli/internal/config"
	"github.com/six78/2-story-points-cli/internal/view/messages"
	"github.com/six78/2-story-points-cli/pkg/protocol"
)

var (
	headerStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA"))
	primaryStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#F0F0F0"))
	secondaryStyle = lipgloss.NewStyle().Foreground(config.ForegroundShadeColor)
	hyperlinkStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#648EF8")).Underline(true)
	errorStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF1744"))
)

const (
	viewHeight = 5
)

type issueFetchedMessage struct {
	url  string
	info *issueInfo
}

type Model struct {
	issue  *protocol.Issue
	issues map[string]*issueInfo

	client  *github.Client
	spinner spinner.Model
}

func New() Model {
	s := spinner.New()
	s.Spinner = spinner.MiniDot

	return Model{
		issue:   nil,
		issues:  make(map[string]*issueInfo),
		client:  github.NewClient(nil),
		spinner: s,
	}
}

func (m Model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case messages.GameStateMessage:
		if msg.State == nil {
			m.issue = nil
			break
		}
		if msg.State.Issues == nil {
			m.issue = nil
			break
		}
		m.issue = msg.State.Issues.Get(msg.State.ActiveIssue)
		if m.issue == nil {
			break
		}
		_, ok := m.issues[m.issue.TitleOrURL]
		if ok {
			break
		}
		cmd = fetchIssue(m.client, m.issue)
		cmds = append(cmds, cmd)
		m.issues[m.issue.TitleOrURL] = nil

	case issueFetchedMessage:
		config.Logger.Debug("issue fetched", zap.Any("msg", msg))
		m.issues[msg.url] = msg.info
	}

	m.spinner, cmd = m.spinner.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.issue == nil {
		return lipgloss.JoinVertical(lipgloss.Center,
			"                                                            ",
			secondaryStyle.Render("No active issue"),
			strings.Repeat("\n", viewHeight-3),
		)
	}

	info := m.issueInfo()
	labelsFirstLine, labelsSecondLine := renderLabelLines(info)

	block := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.JoinVertical(lipgloss.Left,
			headerStyle.Render("Author:"),
			headerStyle.Render("Assignee:"),
		),
		"  ",
		lipgloss.JoinVertical(lipgloss.Left,
			primaryStyle.Render(fmt.Sprintf("%-20s", authorString(info))),
			primaryStyle.Render(fmt.Sprintf("%-20s", assigneeString(info))),
		),
		lipgloss.JoinVertical(lipgloss.Top,
			labelsFirstLine,
			labelsSecondLine,
		),
	)

	return lipgloss.JoinVertical(lipgloss.Left,
		m.renderHeader(),
		m.renderTitle(info),
		"",
		block,
	)
}

func (m *Model) renderHeader() string {
	if m.issue != nil {
		return hyperlinkStyle.Render(m.issue.TitleOrURL)
	}
	return "-"
}

func (m *Model) renderTitle(info *issueInfo) string {
	if m.issue == nil {
		return ""
	}

	if info == nil {
		return secondaryStyle.Render(m.spinner.View() + " fetching title")
	}

	if info.err != nil {
		return errorStyle.Render(fmt.Sprintf("[%s]", info.err.Error()))
	}

	if info.title == nil {
		return secondaryStyle.Render("[empty issue title]")
	}

	return primaryStyle.Render(*info.title) + "  " + secondaryStyle.Render(renderNumber(info))
}

func renderLabels(info *issueInfo) []string {
	if info == nil {
		return []string{}
	}

	var labels []string
	for _, l := range info.labels {
		if l.name == nil {
			continue
		}
		labelName := fmt.Sprintf("[%s]", *l.name)
		labels = append(labels, l.style.Render(labelName))
	}

	return labels
}

func splitLabelsToLines(labels []string) ([]string, []string) {
	if len(labels) < 3 {
		return labels, []string{}
	}

	center := len(labels) / 2
	if len(labels)%2 == 1 {
		center++
	}

	return labels[:center], labels[center:]
}

func joinLabels(labels []string) string {
	return strings.Join(labels, " ")
}

func renderLabelLines(info *issueInfo) (string, string) {
	labels := renderLabels(info)
	labelsFirstLine, labelsSecondLine := splitLabelsToLines(labels)
	return joinLabels(labelsFirstLine), joinLabels(labelsSecondLine)
}

func (m *Model) issueInfo() *issueInfo {
	if m.issue == nil {
		return nil
	}

	info, ok := m.issues[m.issue.TitleOrURL]
	if !ok {
		return nil
	}

	return info
}
