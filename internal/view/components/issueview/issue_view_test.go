package issueview

import (
	"strings"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"

	"github.com/six78/2-story-points-cli/internal/testcommon"
	"github.com/six78/2-story-points-cli/internal/view/messages"
	"github.com/six78/2-story-points-cli/pkg/protocol"
)

func TestHeight(t *testing.T) {
	_ = testcommon.SetupConfigLogger(t)

	model := New()

	var issue protocol.Issue
	err := gofakeit.Struct(&issue)
	require.NoError(t, err)

	var info issueInfo
	err = gofakeit.Struct(&info)
	require.NoError(t, err)

	model, _ = model.Update(messages.GameStateMessage{
		State: &protocol.State{
			Issues:      protocol.IssuesList{&issue},
			ActiveIssue: issue.ID,
		},
	})

	model, _ = model.Update(issueFetchedMessage{
		url:  issue.TitleOrURL,
		info: &info,
	})

	view := model.View()
	lines := strings.Split(view, "\n")
	require.Len(t, lines, viewHeight)
}
