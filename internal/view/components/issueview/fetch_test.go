package issueview

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
)

func TestParseUrl(t *testing.T) {
	owner := gofakeit.LetterN(5)
	repo := gofakeit.LetterN(5)
	issueNumber := gofakeit.Number(1, 1000)

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
			name: "Valid GitHub issue URL",
			url:  fmt.Sprintf("https://github.com/%s/%s/issues/%d", owner, repo, issueNumber),
			expectedResult: &githubIssueRequest{
				owner:  owner,
				repo:   repo,
				number: issueNumber,
			},
			expectedError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			request, err := parseUrl(tc.url)
			require.ErrorIs(t, err, tc.expectedError)
			//require.ErrorAs(t, err, tc.expectedError)
			require.True(t, reflect.DeepEqual(tc.expectedResult, request))
		})
	}
}
