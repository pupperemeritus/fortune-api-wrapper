package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestBuildArgs(t *testing.T) {
	s := NewFortuneService("", zap.NewNop())

	testCases := []struct {
		name     string
		opts     FortuneOptions
		expected []string
	}{
		{
			name:     "No options",
			opts:     FortuneOptions{},
			expected: []string{},
		},
		{
			name:     "Simple flags",
			opts:     FortuneOptions{Short: true, Long: true},
			expected: []string{"-l", "-s"},
		},
		{
			name:     "Length and Pattern",
			opts:     FortuneOptions{Length: 100, Pattern: "test"},
			expected: []string{"-n", "100", "-m", "test"},
		},
		{
			name: "Files and Percentages",
			opts: FortuneOptions{
				Files:       []string{"file1", "file2", "file3"},
				Percentages: []string{"50", "25"},
			},
			expected: []string{"50%", "file1", "25%", "file2", "file3"},
		},
		{
			name: "All options",
			opts: FortuneOptions{
				All:        true,
				ShowCookie: true,
				Equal:      true,
				Long:       true,
				Short:      true,
				IgnoreCase: true,
				Wait:       true,
				Length:     42,
				Pattern:    "life",
				Files:      []string{"hitchhiker"},
			},
			expected: []string{"-a", "-c", "-e", "-l", "-s", "-i", "-w", "-n", "42", "-m", "life", "hitchhiker"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			args := s.buildArgs(tc.opts)
			assert.ElementsMatch(t, tc.expected, args, "Generated args should match expected")
		})
	}
}

func TestParseSearchResults(t *testing.T) {
	s := NewFortuneService("", zap.NewNop())

	testCases := []struct {
		name           string
		output         string
		expectedCount  int
		expectedFirst  string
		expectedSecond string
	}{
		{
			name:          "No results",
			output:        "",
			expectedCount: 0,
		},
		{
			name:          "Single result",
			output:        "This is a fortune.",
			expectedCount: 1,
			expectedFirst: "This is a fortune.",
		},
		{
			name:           "Multiple results",
			output:         "First fortune.\n%\nSecond fortune.", // Correct separator
			expectedCount:  2,
			expectedFirst:  "First fortune.",
			expectedSecond: "Second fortune.",
		},
		{
			name:          "Trailing newlines",
			output:        "A fortune with trailing space.  \n\n",
			expectedCount: 1,
			expectedFirst: "A fortune with trailing space.",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			matches := s.parseSearchResults(tc.output)

			assert.Equal(t, tc.expectedCount, len(matches))
			if tc.expectedCount > 0 {
				assert.Equal(t, tc.expectedFirst, matches[0].Fortune)
			}
			if tc.expectedCount > 1 {
				assert.Equal(t, tc.expectedSecond, matches[1].Fortune)
			}
		})
	}
}
