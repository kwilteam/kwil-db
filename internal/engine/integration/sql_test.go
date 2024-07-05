//go:build pglive

package integration_test

import (
	"context"
	"testing"

	"github.com/kwilteam/kwil-db/internal/engine/execution"
	"github.com/stretchr/testify/require"
)

// the schema deployed here can be found in ./procedure_test.go
func Test_SQL(t *testing.T) {
	type testcase struct {
		name string
		// pre is a sql statement that will be executed before the test
		pre string
		// sql is a sql statement that can be executed
		sql string
		// values is a map of values that can be used in the sql statement
		values map[string]any
		// want is the expected result of the sql statement
		want [][]any
		// err is the expected error, if any
		err error
	}

	tests := []testcase{
		{
			name: "simple select",
			sql:  "SELECT name FROM users",
			want: [][]any{
				{"satoshi"},
				{"wendys_drive_through_lady"},
				{"zeus"},
			},
		},
		{
			name: "select with join",
			sql:  "select u.name, p.content from users u inner join posts p on u.id = p.user_id limit 1;",
			want: [][]any{
				{"satoshi", "goodbye world"},
			},
		},
		{
			name: "aggregate",
			// getting the user and number of posts that have been made by that user
			sql: "select u.name, count(p.id) from users u inner join posts p on u.id = p.user_id group by u.name;",
			want: [][]any{
				{"satoshi", int64(3)},
				{"wendys_drive_through_lady", int64(3)},
				{"zeus", int64(3)},
			},
		},
		{
			name: "compound select",
			sql:  `select name from users union all select name from users`,
			want: [][]any{
				{"satoshi"},
				{"satoshi"},
				{"wendys_drive_through_lady"},
				{"wendys_drive_through_lady"},
				{"zeus"},
				{"zeus"},
			},
		},
		{
			name: "convoluted",
			sql: `select u.name, count(p.id) from (
				select id, name from users union all select '4a67d6ea-7ac8-453c-964e-5a144f9e3004'::uuid, 'hello'
				) u
			left join (
				select id, user_id from posts union all select id, user_id from posts union all select '699e53a3-079c-40a6-b8ae-0d7bb7b40369'::uuid, '4a67d6ea-7ac8-453c-964e-5a144f9e3004'::uuid
			) as p on u.id = p.user_id group by u.name;`,
			want: [][]any{
				{"hello", int64(1)},
				{"satoshi", int64(6)},
				{"wendys_drive_through_lady", int64(6)},
				{"zeus", int64(6)},
			},
		},
		{
			name: "exists and collate",
			sql:  `select exists (select id from users where name = 'SATOSHI' collate nocase)`,
			want: [][]any{
				{true},
			},
		},
		{
			name: "in",
			sql:  `select name from users where name in ('satoshi', 'wendys_drive_through_lady')`,
			want: [][]any{
				{"satoshi"},
				{"wendys_drive_through_lady"},
			},
		},
		{
			name: "like and ilike",
			sql:  `select name from users where name like 's%' or name ilike 'w_Nd%'`,
			want: [][]any{
				{"satoshi"},
				{"wendys_drive_through_lady"},
			},
		},
		{
			name: "unary",
			sql:  `select 22.22=-22.22`,
			want: [][]any{
				{false},
			},
		},
		{
			name: "between",
			sql:  `select name from users where user_num between 2 and 3`,
			want: [][]any{
				{"wendys_drive_through_lady"},
				{"zeus"},
			},
		},
		{
			name: "is, case",
			sql: `select name from (select
				case when name like 's%' then true
				else null end as is_satoshi,
				name
				from users
			) u where is_satoshi is true`,
			want: [][]any{
				{"satoshi"},
			},
		},
		{
			name: "null",
			sql:  `select name from users where user_num is null`,
			want: [][]any{},
		},
		{
			name: "is distinct from",
			sql:  `select name from users where user_num is distinct from 2`,
			want: [][]any{
				{"satoshi"},
				{"zeus"},
			},
		},
		{
			name: "insert with conflict",
			// this will conflict on user_num = 1
			pre: `insert into users (id, name, user_num, wallet_address) values ('4a67d6ea-7ac8-453c-964e-5a144f9e3004'::uuid, 'hello', 1, '0xa'), ('4a67d6ea-7ac8-453c-964e-5a144f9e3005'::uuid, 'hello2', 4, '0xb')
			on conflict (user_num) do update set name = 'hello3', user_num = excluded.user_num*10`,
			sql: `select user_num from users where name like 'hello%'`,
			want: [][]any{
				{int64(10)},
				{int64(4)},
			},
		},
		{
			name: "update with subquery",
			pre: `update users set wallet_address = 'hello' where id in (
				// hack here since we can't use aggregates in where clauses
				select id from users u inner join (select count(*) as count, user_id from posts group by user_id) p on u.id = p.user_id where count >= 3
			)`,
			sql: `select count(*) from users where wallet_address = 'hello'`,
			want: [][]any{
				{int64(3)},
			},
		},
		{
			name: "update from",
			pre:  `update posts p set content = u.name from users u where p.user_id = u.id`,
			sql:  `select distinct content from posts`,
			want: [][]any{
				{"satoshi"},
				{"wendys_drive_through_lady"},
				{"zeus"},
			},
		},
		{
			name: "delete",
			pre:  `delete from users where name = 'satoshi'`,
			sql:  `select name from users`,
			want: [][]any{
				{"wendys_drive_through_lady"},
				{"zeus"},
			},
		},
		{
			name: "select constant",
			sql:  "select 1",
			want: [][]any{
				{int64(1)},
			},
		},
		{
			// this is a regression test for a bug introduced
			// in v0.8
			name: "values",
			sql:  "select $id",
			values: map[string]any{
				"id": "4a67d6ea-7ac8-453c-964e-5a144f9e3004",
			},
			want: [][]any{
				{"4a67d6ea-7ac8-453c-964e-5a144f9e3004"},
			},
		},
		{
			name: "inferred type - failure",
			sql:  "select $id is null",
			values: map[string]any{
				"id": "4a67d6ea-7ac8-453c-964e-5a144f9e3004",
			},
			err: execution.ErrCannotInferType,
		},
		{
			name: "inferred type - success",
			sql:  "select $id::text is null",
			values: map[string]any{
				"id": "4a67d6ea-7ac8-453c-964e-5a144f9e3004",
			},
			want: [][]any{{false}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			global, db, err := setup(t)
			require.NoError(t, err)
			defer cleanup(t, db)

			ctx := context.Background()

			tx, err := db.BeginOuterTx(ctx)
			require.NoError(t, err)
			defer tx.Rollback(ctx)

			// deploy schema
			dbid := deployAndSeed(t, global, tx)

			// if there is a pre statement, execute it
			if tt.pre != "" {
				_, err := global.Execute(ctx, tx, dbid, tt.pre, nil)
				require.NoError(t, err)
			}

			// execute sql
			res, err := global.Execute(ctx, tx, dbid, tt.sql, tt.values)
			if tt.err != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.err)
				return
			}
			require.NoError(t, err)

			require.Len(t, res.Rows, len(tt.want))
			for i, row := range res.Rows {
				require.Len(t, row, len(tt.want[i]))
				for j, col := range row {
					require.Equal(t, tt.want[i][j], col)
				}
			}

		})
	}
}
