package sql

import "testing"

func TestParseRawSQL_banRules(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool // TODO: use custom error type
	}{
		// non-deterministic functions
		{"date function", "select date('now')", true},
		{"time function", "select time('now')", true},
		{"datetime function", "select datetime('now')", true},
		{"julianday function", "select julianday('now')", true},
		{"unixepoch function", "select unixepoch('now')", true},
		{"random function", "select random()", true},
		{"randomblob function", "select randomblob(10)", true},
		// non-deterministic keywords
		{"current_date", "select current_date", true},
		{"current_time", "select current_time", true},
		{"current_timestamp", "select current_timestamp", true},
		// joins
		{"natural join", "select * from users natural join posts", true},
		{"cross join", "select * from users cross join posts", true},
		{"cartesian join 1", "select * from users, posts", true},
		// join without any condition
		{"cartesian join 2 1", "select * from users left join posts", true},
		{"cartesian join 2 2", "select * from users inner join posts", true},
		{"cartesian join 2 3", "select * from users join posts", true},

		// join with condition
		//{"join with binary cond", "select * from users join posts on users.id = posts.user_id", false},
		//{"join with multi level binary cond", "select * from users join posts on a=(b+c)", false},
		{"join with unary cond", "select * from users join posts on not a", true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ParseRawSQL(test.input, 1, false)
			if err != nil && !test.wantError {
				t.Errorf("ParseRawSQL() failed with expected error: %s", err)
			}
		})
	}
}
