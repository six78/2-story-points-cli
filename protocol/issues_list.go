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

func (l IssuesList) GetNextIssueToDeal(finishedIssueID IssueID) IssueID {
	finishedIssueIndex := -1
	for i, issue := range l {
		if issue.ID != finishedIssueID {
			continue
		}
		finishedIssueIndex = i
	}
	for i := finishedIssueIndex; i < len(l); i++ {
		if l[i].Result == nil {
			return l[i].ID
		}
	}
	for i := 0; i < finishedIssueIndex; i++ {
		if l[i].Result == nil {
			return l[i].ID
		}
	}
	return ""
}
