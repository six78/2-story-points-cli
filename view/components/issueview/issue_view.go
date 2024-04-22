package issueview

import (
	"context"
	"fmt"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/go-github/v61/github"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"net/url"
	"strconv"
	"strings"
	"time"
	"waku-poker-planning/config"
	"waku-poker-planning/protocol"
	"waku-poker-planning/view/messages"
)

var (
	errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))
	//defaultStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))
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

type issueInfo struct {
	err   error
	title *string
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
		config.Logger.Debug("<<< issue fetched",
			zap.Any("msg", msg),
		)
		m.issues[msg.url] = msg.info
	}

	m.spinner, cmd = m.spinner.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	return fmt.Sprintf("Issue:  %s\n        %s",
		m.renderRow1(),
		m.renderRow2(),
	)
}

func (m *Model) renderRow1() string {
	if m.issue != nil {
		return m.issue.TitleOrURL
	}
	return "-"
}

func (m *Model) renderRow2() string {
	if m.issue == nil {
		return ""
	}

	info, ok := m.issues[m.issue.TitleOrURL]
	if !ok {
		return ""
	}

	if info == nil {
		return errorStyle.Render(m.spinner.View() + " fetching title")
	}

	if info.err != nil {
		return errorStyle.Render(fmt.Sprintf("[%s]", info.err.Error()))
	}

	if info.title == nil {
		return errorStyle.Render("[empty issue title]")
	}

	return *info.title
}

type githubIssueRequest struct {
	owner  string
	repo   string
	number int
}

func parseUrl(input string) (*githubIssueRequest, error) {
	u, err := url.Parse(input)
	if err != nil {
		return nil, nil
	}
	if u.Host != "github.com" {
		return nil, errors.New("only github links are supported")
	}
	path := strings.Split(u.Path, "/")
	if len(path) != 5 {
		return nil, errors.New("invalid github issue link")
	}

	issueNumber, err := strconv.Atoi(path[4])
	if err != nil {
		return nil, errors.New("invalid github issue number")
	}

	return &githubIssueRequest{
		owner:  path[1],
		repo:   path[2],
		number: issueNumber,
	}, nil
}

func fetchIssue(client *github.Client, input *protocol.Issue) tea.Cmd {
	return func() tea.Msg {
		if input == nil {
			return nil
		}
		request, err := parseUrl(input.TitleOrURL)
		if err != nil {
			return issueFetchedMessage{
				url: input.TitleOrURL,
				info: &issueInfo{
					err: err,
				},
			}
		}
		if request == nil {
			return nil
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		issue, _, err := client.Issues.Get(ctx, request.owner, request.repo, request.number)
		if err != nil {
			return issueFetchedMessage{
				url: input.TitleOrURL,
				info: &issueInfo{
					err: errors.New("failed to fetch github issue"),
				},
			}
		}

		return issueFetchedMessage{
			url: input.TitleOrURL,
			info: &issueInfo{
				err:   nil,
				title: issue.Title,
			},
		}
	}
}
