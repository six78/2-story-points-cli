package cursor

import (
	"fmt"
	"github.com/stretchr/testify/suite"
	"testing"
)

func TestCursor(t *testing.T) {
	suite.Run(t, new(Suite))
}

type Suite struct {
	suite.Suite
}

func (s *Suite) TestDecrementCursor() {
	testCases := []struct {
		given    int
		expected int
	}{
		{
			given:    0,
			expected: 0,
		},
		{
			given:    1,
			expected: 0,
		},
		{
			given:    2,
			expected: 1,
		},
	}

	for _, tc := range testCases {
		test := func(*testing.T) {
			model := New(true, true)
			model.SetRange(0, 2)
			model.decrementCursor(tc.given)
			s.Require().Equal(tc.expected, model.Position())
		}
		testName := fmt.Sprintf("test decrement %d", tc.given)
		s.T().Run(testName, test)
	}
}

func (s *Suite) TestIncrementCursor() {
	testCases := []struct {
		given    int
		expected int
	}{
		{
			given:    0,
			expected: 1,
		},
		{
			given:    1,
			expected: 2,
		},
		{
			given:    2,
			expected: 2,
		},
	}

	for _, tc := range testCases {
		test := func(*testing.T) {
			model := New(true, true)
			model.SetRange(0, 2)
			model.incrementCursor(tc.given)
			s.Require().Equal(tc.expected, model.Position())
		}
		testName := fmt.Sprintf("test increment %d", tc.given)
		s.T().Run(testName, test)
	}
}
