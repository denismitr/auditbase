package mysql

import (
	"context"
	"database/sql"
	"github.com/denismitr/auditbase/internal/db"
	"github.com/denismitr/auditbase/internal/utils/logger"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

const MySQL8 = "mysql8"

type Database struct {
	conn *sqlx.DB
	lg   logger.Logger
}

func NewDatabase(conn *sqlx.DB, lg logger.Logger) *Database {
	return &Database{
		conn:  conn,
		lg:    lg,
	}
}

type Tx struct {
	mysqlTx *sqlx.Tx
	lg      logger.Logger
}

var _ db.Tx = (*Tx)(nil)

func (db *Database) ReadOnly(ctx context.Context, cb db.TxCallback) (interface{}, error) {
	mysqlTx, err := db.conn.BeginTxx(ctx, &sql.TxOptions{ReadOnly: true, Isolation: sql.LevelSerializable})
	if err != nil {
		return nil, errors.Wrap(err, "could not start read only Tx")
	}

	result, err := cb(ctx, &Tx{mysqlTx: mysqlTx, lg: db.lg});
	if err != nil {
		if rbErr := mysqlTx.Rollback(); rbErr != nil {
			return nil, errors.Wrap(err, rbErr.Error())
		}

		return nil, err
	}

	if err := mysqlTx.Commit(); err != nil {
		return nil, errors.Wrap(err, "could not commit read only Tx")
	}

	return result, nil
}

func (db *Database) ReadWrite(ctx context.Context, cb db.TxCallback) (interface{}, error) {
	mysqlTx, err := db.conn.BeginTxx(ctx, &sql.TxOptions{ReadOnly: false, Isolation: sql.LevelSerializable})
	if err != nil {
		return nil, errors.Wrap(err, "could not start read write Tx")
	}

	result, err := cb(ctx, &Tx{mysqlTx: mysqlTx, lg: db.lg});
	if err != nil {
		if rbErr := mysqlTx.Rollback(); rbErr != nil {
			return nil, errors.Wrap(err, rbErr.Error())
		}

		return nil, err
	}

	if err := mysqlTx.Commit(); err != nil {
		return nil, errors.Wrap(err, "could not commit read write Tx")
	}

	return result, nil
}

func (tx *Tx) Entities() db.EntityRepository {
	return &EntityRepository{Tx: tx}
}

func (tx *Tx) EntityTypes() db.EntityTypeRepository {
	return &EntityTypeRepository{Tx: tx}
}

func (tx *Tx) Microservices() db.MicroserviceRepository {
	return &MicroserviceRepository{Tx: tx}
}

func (tx *Tx) Actions() db.ActionRepository {
	return &ActionRepository{Tx: tx}
}
