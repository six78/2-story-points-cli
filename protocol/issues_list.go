package protocol

type IssuesList []*Issue

func (l *IssuesList) Get(id IssueID) *Issue {
	for _, issue := range *l {
		if issue.ID == id {
			return issue
		}
	}

	return nil
}
