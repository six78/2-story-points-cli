package issueview

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/go-github/v61/github"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	mock "github.com/six78/2-story-points-cli/internal/view/components/issueview/mock"
	"github.com/six78/2-story-points-cli/pkg/protocol"
)

func fakeIssueURL() (*githubIssueRequest, string) {
	request := &githubIssueRequest{
		owner:  gofakeit.LetterN(5),
		repo:   gofakeit.LetterN(5),
		number: gofakeit.Number(1, 1000),
	}
	url := fmt.Sprintf("https://github.com/%s/%s/issues/%d",
		request.owner,
		request.repo,
		request.number)
	return request, url
}

func TestParseUrl(t *testing.T) {
	expectedRequest, validURL := fakeIssueURL()

	testCases := []struct {
		name           string
		url            string
		expectedResult *githubIssueRequest
		expectedError  error
	}{
		{
			name:           "Not a URL",
			url:            "://not-a-url",
			expectedResult: nil,
			expectedError:  nil, // We want not-urls to be silently ignored
		},
		{
			name:           "Not a GitHub URL",
			url:            "https://example.com",
			expectedResult: nil,
			expectedError:  errOnlyGithubIssuesUnfurled,
		},
		{
			name:           "Invalid GitHub issue link",
			url:            "https://github.com/url/path/should/have/exactly/five/parts",
			expectedResult: nil,
			expectedError:  errInvalidGithubIssueLink,
		},
		{
			name:           "Invalid GitHub issue number",
			url:            "https://github.com/owner/repo/issues/issue-number-is-not-a-number",
			expectedResult: nil,
			expectedError:  errInvalidGithubIssueNumber,
		},
		{
			name:           "Valid GitHub issue URL",
			url:            validURL,
			expectedResult: expectedRequest,
			expectedError:  nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			request, err := parseUrl(tc.url)
			require.ErrorIs(t, err, tc.expectedError)
			require.True(t, reflect.DeepEqual(tc.expectedResult, request))
		})
	}
}

// TestParseError
// TestFetchError

func TestFetchIssue(t *testing.T) {
	ctrl := gomock.NewController(t)
	client, issue, ghIssue := setupClientWithGeneratedIssue(ctrl)

	cmd := fetchIssue(client, issue)
	require.NotNil(t, cmd)

	msg := cmd()
	require.NotNil(t, msg)

	issueMessage := msg.(issueFetchedMessage)
	require.NoError(t, issueMessage.info.err)
	require.Equal(t, *ghIssue.Number, *issueMessage.info.number)
	require.Equal(t, *ghIssue.Title, *issueMessage.info.title)
	require.Equal(t, *ghIssue.User.Login, *issueMessage.info.author)
	require.Equal(t, *ghIssue.Assignee.Login, *issueMessage.info.assignee)
	require.Len(t, issueMessage.info.labels, len(ghIssue.Labels))
	for i, label := range ghIssue.Labels {
		require.Equal(t, *label.Name, *issueMessage.info.labels[i].name)
	}
}

func TestParseURLError(t *testing.T) {
	issue := &protocol.Issue{
		ID:         protocol.IssueID(gofakeit.UUID()),
		TitleOrURL: "https://example.com",
	}

	cmd := fetchIssue(nil, issue)
	require.NotNil(t, cmd)

	msg := cmd()
	require.NotNil(t, msg)

	issueMessage := msg.(issueFetchedMessage)
	require.Equal(t, issueMessage.url, issue.TitleOrURL)
	require.ErrorIs(t, issueMessage.info.err, errOnlyGithubIssuesUnfurled)
}

func TestNotURL(t *testing.T) {
	issue := &protocol.Issue{
		ID:         protocol.IssueID(gofakeit.UUID()),
		TitleOrURL: "://not-a-url",
	}

	cmd := fetchIssue(nil, issue)
	require.NotNil(t, cmd)

	msg := cmd()
	require.Nil(t, msg)
}

func TestFetchError(t *testing.T) {
	issue, _, urlInfo := generateIssue()

	ctrl := gomock.NewController(t)
	client := mock.NewMockGithubIssueService(ctrl)

	client.EXPECT().
		Get(gomock.Any(), urlInfo.owner, urlInfo.repo, urlInfo.number).
		Return(nil, nil, fmt.Errorf("error")).
		Times(1)

	cmd := fetchIssue(client, issue)
	require.NotNil(t, cmd)

	msg := cmd()
	require.NotNil(t, msg)

	issueMessage := msg.(issueFetchedMessage)
	require.Equal(t, issueMessage.url, issue.TitleOrURL)
	require.ErrorIs(t, issueMessage.info.err, errGithubIssueFetchFailed)
}

func TestFetchNil(t *testing.T) {
	cmd := fetchIssue(nil, nil)
	require.NotNil(t, cmd)

	msg := cmd()
	require.Nil(t, msg)
}

func generateIssue() (*protocol.Issue, *github.Issue, *githubIssueRequest) {
	urlInfo, url := fakeIssueURL()

	issue := &protocol.Issue{
		ID:         protocol.IssueID(gofakeit.UUID()),
		TitleOrURL: url,
	}

	// Generate issue
	title := gofakeit.LetterN(10)
	author := gofakeit.LetterN(5)
	assignee := gofakeit.LetterN(5)
	ghIssue := &github.Issue{
		Number: &urlInfo.number,
		Title:  &title,
		User: &github.User{
			Login: &author,
		},
		Assignee: &github.User{
			Login: &assignee,
		},
	}
	for i := 0; i < gofakeit.Number(1, 3); i++ {
		labelLength := uint(gofakeit.Number(2, 20))
		name := gofakeit.LetterN(labelLength)
		label := &github.Label{
			Name: &name,
		}
		if i == 0 {
			color := gofakeit.HexColor()
			label.Color = &color
		}
		ghIssue.Labels = append(ghIssue.Labels, label)
	}

	return issue, ghIssue, urlInfo
}

func setupClientWithGeneratedIssue(ctrl *gomock.Controller) (GithubIssueService, *protocol.Issue, *github.Issue) {
	issue, ghIssue, urlInfo := generateIssue()

	client := mock.NewMockGithubIssueService(ctrl)
	client.EXPECT().
		Get(gomock.Any(), urlInfo.owner, urlInfo.repo, urlInfo.number).
		Return(ghIssue, nil, nil).
		Times(1)

	return client, issue, ghIssue
}
