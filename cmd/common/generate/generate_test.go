package generate

import "testing"

func Test_Links(t *testing.T) {
	type testcase struct {
		currentDir string
		link       string
		expected   string
	}

	testcases := []testcase{
		{
			currentDir: "cli/cmd1/cmd2",
			link:       "cli_cmd1_cmd2",
			expected:   ".",
		},
		{
			currentDir: "cli/cmd1/cmd2",
			link:       "cli_cmd1_cmd2_cmd3",
			expected:   "./cmd3",
		},
		{
			currentDir: "cli/cmd1/cmd2",
			link:       "cli_cmd1",
			expected:   "../",
		},
		{
			currentDir: "./kwil-cli/utils",
			link:       "kwil-cli_utils_query-tx.md",
			expected:   "./query-tx",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.currentDir+"_"+tc.link, func(t *testing.T) {
			link := linkHandler(tc.currentDir)(tc.link)
			if link != tc.expected {
				t.Errorf("expected %s, got %s", tc.expected, link)
			}
		})
	}
}
