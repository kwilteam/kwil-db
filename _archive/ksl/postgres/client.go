package postgres

import (
	"context"
	"fmt"
	"hash/fnv"
	"io"
	"ksl/sqldriver"
	"ksl/sqlmigrate"
	"ksl/sqlschema"
	"ksl/sqlx"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Client struct {
	sqldriver.ExecQuerier
	sqlmigrate.Planner
	sqlmigrate.Differ
	sqlmigrate.Describer
	sqldriver.Executor
}

func NewClient(db sqldriver.ExecQuerier) *Client {
	describer := Describer{Conn: db}

	return &Client{
		ExecQuerier: db,
		Planner:     Planner{},
		Differ:      sqlmigrate.NewDiffer(Backend{}),
		Describer:   describer,
		Executor:    &Executor{Describer: describer, Conn: db},
	}
}

func (c *Client) Close() error {
	if closer, ok := c.ExecQuerier.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

func (c *Client) Lock(ctx context.Context, name string, timeout time.Duration) (sqldriver.UnlockFunc, error) {
	conn, err := sqldriver.SingleConn(ctx, c.ExecQuerier)
	if err != nil {
		return nil, err
	}

	h := fnv.New32()
	h.Write([]byte(name))
	id := h.Sum32()

	if err := lockAcquire(ctx, conn, id, timeout); err != nil {
		conn.Close()
		return nil, err
	}

	return func() error {
		defer conn.Close()
		rows, err := conn.QueryContext(ctx, "SELECT pg_advisory_unlock($1)", id)
		if err != nil {
			return err
		}

		switch released, err := sqlx.ScanNullBool(rows); {
		case err != nil:
			return err
		case !released.Valid || !released.Bool:
			return fmt.Errorf("postgres: failed releasing lock %d", id)
		}
		return nil
	}, nil
}

func (c *Client) ServerVersion(ctx context.Context) (string, error) {
	var version string
	rows, err := c.ExecQuerier.QueryContext(ctx, "SELECT version()")
	if err != nil {
		return "", err
	}
	if err := sqlx.ScanOne(rows, &version); err != nil {
		return "", err
	}

	return version, nil
}

func (c *Client) ApplyMigration(ctx context.Context, plan sqlmigrate.MigrationPlan) error {
	for _, stmt := range plan.Statements {
		for _, step := range stmt.Steps {
			if _, err := c.ExecContext(ctx, step.Cmd, step.Args...); err != nil {
				if step.Comment != "" {
					err = fmt.Errorf("%s: %w", step.Comment, err)
				}
				return err
			}
		}
	}
	return nil
}

func (c *Client) PlanMigration(ctx context.Context, before, after sqlschema.Database) (sqlmigrate.MigrationPlan, error) {
	steps, err := c.Diff(before, after)
	if err != nil {
		return sqlmigrate.MigrationPlan{}, err
	}

	migration := sqlmigrate.Migration{
		Before:  before,
		After:   after,
		Changes: steps,
	}

	return c.Plan(migration)
}
