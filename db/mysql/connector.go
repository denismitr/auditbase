package mysql

import (
	"context"
	"github.com/denismitr/auditbase/utils/logger"
	"github.com/denismitr/auditbase/utils/retry"
	"github.com/jmoiron/sqlx"
	"time"
)

func ConnectAndMigrate(ctx context.Context, lg logger.Logger, dsn string, maxOpenConnection, maxIdleConnections int) (*sqlx.DB, error) {
	var conn *sqlx.DB
	if err := retry.Incremental(ctx, 2 * time.Second, 100, func(attempt int) (err error) {
		conn, err = sqlx.Connect("mysql", dsn)
		if err != nil {
			lg.Debugf("could not connect to DB on attempt %d", attempt)
			return retry.Error(err, attempt)
		}

		if _, err = conn.QueryxContext(ctx, "select 1"); err != nil {
			lg.Debugf("could not ping DB connection on attempt %d", attempt)
			return retry.Error(err, attempt)
		}

		lg.Debugf("connection with DB established")
		return nil
	}); err != nil {
		return nil, err
	}

	conn.SetMaxOpenConns(maxOpenConnection)
	conn.SetMaxIdleConns(maxIdleConnections)

	if err := Migrator(conn, lg).Up(); err != nil {
		return nil, err
	}

	return conn, nil
}
