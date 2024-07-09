package issueview

import (
	"context"
	"net/url"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/pkg/errors"

	"github.com/six78/2-story-points-cli/internal/config"
	"github.com/six78/2-story-points-cli/pkg/protocol"
)

var (
	errOnlyGithubIssuesUnfurled = errors.New("only github issues can be unfurled")
	errInvalidGithubIssueLink   = errors.New("invalid github issue link")
	errInvalidGithubIssueNumber = errors.New("invalid github issue number")
	errGithubIssueFetchFailed   = errors.New("failed to fetch github issue")
)

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
		return nil, errOnlyGithubIssuesUnfurled
	}
	path := strings.Split(u.Path, "/")
	if len(path) != 5 {
		return nil, errInvalidGithubIssueLink
	}

	issueNumber, err := strconv.Atoi(path[4])
	if err != nil {
		return nil, errInvalidGithubIssueNumber
	}

	return &githubIssueRequest{
		owner:  path[1],
		repo:   path[2],
		number: issueNumber,
	}, nil
}

func fetchIssue(client GithubIssueService, input *protocol.Issue) tea.Cmd {
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

		issue, _, err := client.Get(ctx, request.owner, request.repo, request.number)
		if err != nil {
			return issueFetchedMessage{
				url: input.TitleOrURL,
				info: &issueInfo{
					err: errGithubIssueFetchFailed,
				},
			}
		}

		labels := make([]labelInfo, len(issue.Labels))
		for i, label := range issue.Labels {
			labels[i].name = label.Name
			labels[i].style = labelStyle(label.Color)
		}

		msg := issueFetchedMessage{
			url: input.TitleOrURL,
			info: &issueInfo{
				err:    nil,
				number: issue.Number,
				title:  issue.Title,
				labels: labels,
			},
		}

		if issue.User != nil {
			msg.info.author = issue.User.Login
		}

		if issue.Assignee != nil {
			msg.info.assignee = issue.Assignee.Login
		}

		return msg
	}
}

func labelStyle(input *string) lipgloss.Style {
	if input == nil {
		return lipgloss.NewStyle().Foreground(config.ForegroundShadeColor)
	}

	color := lipgloss.Color("#" + *input)
	dark := colorIsDark(color)

	if lipgloss.DefaultRenderer().HasDarkBackground() == dark {
		return lipgloss.NewStyle().Background(color)
	}

	return lipgloss.NewStyle().Foreground(color)
}

func colorIsDark(color lipgloss.Color) bool {
	renderer := lipgloss.DefaultRenderer()
	c := renderer.ColorProfile().Color(string(color))
	rgb := termenv.ConvertToRGB(c)
	//_, _, lightness := rgb.Hsl()
	perceivedLightness := 0.2126*rgb.R + 0.7152*rgb.G + 0.0722*rgb.B
	return perceivedLightness < 0.453
}
