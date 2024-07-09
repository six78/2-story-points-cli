package issueview

import (
	"fmt"
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

func TestSplitLabelsToLines(t *testing.T) {
	testCases := []struct {
		labelsSizes   []uint
		expectedSplit int
	}{
		{labelsSizes: []uint{1}, expectedSplit: 1},
		{labelsSizes: []uint{1, 1}, expectedSplit: 1},
		{labelsSizes: []uint{1, 1, 1}, expectedSplit: 1},
		{labelsSizes: []uint{1, 1, 1, 1}, expectedSplit: 2},
		{labelsSizes: []uint{1, 1, 1, 1}, expectedSplit: 2},
		{labelsSizes: []uint{10, 1, 1, 1}, expectedSplit: 1}, // at least one issue at first line
		{labelsSizes: []uint{1, 1, 1, 10}, expectedSplit: 3},
	}

	for _, tc := range testCases {
		name := fmt.Sprint(tc.labelsSizes)
		t.Run(name, func(t *testing.T) {
			labels := make([]labelInfo, len(tc.labelsSizes))
			for i := range labels {
				label := gofakeit.LetterN(tc.labelsSizes[i])
				labels[i].name = &label
			}

			splitIndex := splitLabelsToLines(labels)
			require.Equal(t, tc.expectedSplit, splitIndex)
		})
	}
}
