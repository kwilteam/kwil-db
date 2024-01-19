package pg

import (
	"context"

	"github.com/jackc/pgx/v5"
)

const (
	sqlCreateFuncError = `CREATE OR REPLACE FUNCTION error(msg text)
		RETURNS void AS $$
		BEGIN
			RAISE EXCEPTION '%', msg;
		END;
		$$ LANGUAGE plpgsql;`
)

func ensureErrorPLFunc(ctx context.Context, conn *pgx.Conn) error {
	_, err := conn.Exec(ctx, sqlCreateFuncError)
	return err
}
