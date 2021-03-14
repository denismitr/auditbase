package mysql

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/denismitr/auditbase/db"
	"github.com/denismitr/auditbase/model"
	"github.com/denismitr/auditbase/utils/validator"
	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/mysql"
	"github.com/pkg/errors"
	"time"
)

type StringInterfaceMap map[string]interface{}

type ActionRepository struct {
	*Tx
}

type actionRecord struct {
	ID             string         `db:"id"`
	ParentEventID  sql.NullString `db:"parent_event_id"`
	Hash           string         `db:"hash"`
	ActorEntityID  sql.NullString `db:"actor_entity_id"`
	TargetEntityID sql.NullString `db:"target_entity_id"`
	Name           string         `db:"name"`
	Details        *string         `db:"details"`
	Delta          *string         `db:"delta"`
	EmittedAt      time.Time      `db:"emitted_at"`
	RegisteredAt   time.Time      `db:"registered_at"`
}

var _ db.ActionRepository = (*ActionRepository)(nil)

func (r *ActionRepository) Create(ctx context.Context, action *model.Action) (*model.Action, error) {
	q, args, err := createActionQuery(action)
	if err != nil {
		panic(fmt.Sprintf("how could createActionQuery func have failed? %s", err.Error()))
	}

	stmt, err := r.mysqlTx.PreparexContext(ctx, q)
	if err != nil {
		return nil, errors.Wrapf(err, "could not prepare query %s", q)
	}

	if _, err := stmt.ExecContext(ctx, args...); err != nil {
		return nil, errors.Wrap(err, "could not create action")
	}

	return r.FirstByID(ctx, action.ID)
}

func (r *ActionRepository) CountAll(ctx context.Context) (int, error) {
	q := `select count(*) as cnt from actions`

	stmt, err := r.mysqlTx.PreparexContext(ctx, q)
	if err != nil {
		return 0, errors.Wrap(err, "could not prepare count all actions")
	}

	var count int
	row := stmt.QueryRowContext(ctx)
	switch err := row.Scan(&count); err {
	case sql.ErrNoRows:
		return 0, nil
	case nil:
		return count, nil
	default:
		return 0, errors.Wrap(err, "could not count actions")
	}
}

func (r *ActionRepository) Select(
	ctx context.Context,
	c *db.Cursor,
	f *db.Filter,
) (*model.ActionCollection, error) {
	sQ, err := selectActionsQuery(c, f)
	if err != nil {
		panic(fmt.Sprintf("how could selectActionsQuery func fail with %s", err))
	}

	selectStmt, err := r.mysqlTx.PreparexContext(ctx, sQ.selectSQL)
	if err != nil {
		return nil, errors.Wrap(err, "could not prepare select actions query")
	}

	countStmt, err := r.mysqlTx.PreparexContext(ctx, sQ.countSQL)
	if err != nil {
		return nil, errors.Wrap(err, "could not prepare select actions query")
	}

	var total int
	if err := countStmt.GetContext(ctx, &total); err != nil {
		return nil, errors.Wrap(err, "could not execute count actions query")
	}

	var ars []actionRecord
	if err := selectStmt.GetContext(ctx, &ars); err != nil {
		return nil, errors.Wrap(err, "could not execute select actions query")
	}

	return mapActionRecordsToCollection(ars, total, c.Page, c.PerPage), nil
}

func selectActionsQuery(c *db.Cursor, f *db.Filter) (*selectQuery, error) {
	dialectCq := goqu.Dialect(MySQL8)
	dialectSq := goqu.Dialect(MySQL8)

	countQ := dialectCq.
		From("actions").
		Select(goqu.L("count(*)").As("cnt"))

	q := dialectSq.Select(
		goqu.L("bin_to_uuid(id)").As("id"),
		"name", "hash",
		goqu.L("ifnull(parent_id, bin_to_uuid(parent_id))").As("parent_id"),
		goqu.L("ifnull(actor_entity_id, bin_to_uuid(actor_entity_id))").As("actor_entity_id"),
		goqu.L("ifnull(target_entity_id, bin_to_uuid(target_entity_id))").As("target_entity_id"),
		"emitted_at", "registered_at",
	).From("actions")

	if f.Has("parentId") {
		exp := goqu.L(`parent_id = uuid_to_bin(?)`, f.StringOrDefault("parentId", ""))
		countQ = countQ.Where(exp)
		q = q.Where(exp)
	}

	if f.Has("actorEntityId") {
		exp := goqu.L(`actor_entity_id = uuid_to_bin(?)`, f.StringOrDefault("actor_entity_id", ""))
		countQ = countQ.Where(exp)
		q = q.Where(exp)
	}

	if f.Has("targetEntityId") {
		exp := goqu.L(`target_entity_id = uuid_to_bin(?)`, f.StringOrDefault("target_entity_id", ""))
		countQ = countQ.Where(exp)
		q = q.Where(exp)
	}

	if c.Sort.Has("name") {
		expr := goqu.I("name")
		if c.Sort.GetOrDefault("name", db.DESCOrder) == db.ASCOrder {
			q = q.Order(expr.Asc())
		} else {
			q = q.Order(expr.Desc())
		}
	} else {
		expr := goqu.I("registered_at")
		if c.Sort.GetOrDefault("registered_at", db.DESCOrder) == db.ASCOrder {
			q = q.Order(expr.Asc())
		} else {
			q = q.Order(expr.Desc())
		}
	}

	q = q.Limit(c.PerPage)
	q = q.Offset(c.Offset())

	sQ := selectQuery{}
	if query, args, err := q.ToSQL(); err != nil {
		return nil, errors.Wrap(err, "invalid select SQL for entities")
	} else {
		sQ.selectSQL = query
		sQ.selectArgs = args
	}

	if query, args, err := countQ.ToSQL(); err != nil {
		return nil, errors.Wrap(err, "invalid count SQL for entities")
	} else {
		sQ.countSQL = query
		sQ.countArgs = args
	}

	return &sQ, nil
}

func (r *ActionRepository) FirstByID(ctx context.Context, ID model.ID) (*model.Action, error) {
	q, args, err := firstActionByIDQuery(ID)
	if err != nil {
		panic("how could firstActionByIDQuery func fail?")
	}

	stmt, err := r.mysqlTx.PreparexContext(ctx, q)
	if err != nil {
		return nil, errors.Wrap(err, "could not prepare firstByID action query")
	}

	var ar actionRecord

	if err := stmt.GetContext(ctx, &ar, args...); err != nil {
		if err == sql.ErrNoRows {
			return nil, db.ErrNotFound
		}

		return nil, errors.Wrap(err, "could not get action with ID [%s]")
	}

	return mapActionRecordToModel(ar), nil
}

func (r *ActionRepository) Delete(ctx context.Context, ID model.ID) error {
	q, args, err := deleteActionQuery(ID)
	if err != nil {
		panic("how could deleteActionQuery fail?")
	}

	stmt, err := r.mysqlTx.PreparexContext(ctx, q)
	if err != nil {
		return errors.Wrap(err, "could not prepare delete action query")
	}

	if _, err := stmt.ExecContext(ctx, args...); err != nil {
		if err == sql.ErrNoRows {
			return errors.Wrapf(db.ErrActionNotFound, "id [%s]", ID.String())
		}

		return errors.Wrapf(err, "could not delete action with ID [%s]", ID.String())
	}

	return nil
}

func (r *ActionRepository) Names(ctx context.Context) ([]string, error) {
	q, _, err := actionNamesQuery()
	if err != nil {
		panic(fmt.Sprintf("how could actionNamesQuery func fail? %s", err))
	}

	stmt, err := r.mysqlTx.PreparexContext(ctx, q)
	if err != nil {
		return nil, errors.Wrap(err, "could not prepare delete action query")
	}

	var names []string
	if err := stmt.SelectContext(ctx, &names); err != nil {
		return nil, errors.Wrap(err, "could not select action names")
	}

	return names, nil
}

func actionNamesQuery() (string, []interface{}, error) {
	dialect := goqu.Dialect(MySQL8)

	return dialect.From("actions").Select(
		goqu.L("distinct name"),
	).ToSQL()
}

func deleteActionQuery(ID model.ID) (string, []interface{}, error) {
	if !validator.IsUUID4(ID.String()) {
		panic("how can action id not be a valid UUID4?")
	}

	return sq.Delete("actions").Where("id = uuid_to_bin(?)", ID.String()).ToSql()
}

func firstActionByIDQuery(ID model.ID) (string, []interface{}, error) {
	if !validator.IsUUID4(ID.String()) {
		panic("how can action id not be a valid UUID4?")
	}

	dialect := goqu.Dialect(MySQL8)

	return dialect.From("actions").Select(
		goqu.L("bin_to_uuid(id)").As("id"),
		"hash", "name", "details", "delta", "emitted_at", "registered_at",
	).Where(
		goqu.L("id = uuid_to_bin(?)", ID.String()),
	).Limit(1).ToSQL()
}

func createActionQuery(action *model.Action) (string, []interface{}, error) {
	dialect := goqu.Dialect(MySQL8)

	row := goqu.Record{}

	if action.ID == "" {
		panic("how can action ID be empty?")
	}

	row["id"] = goqu.L("uuid_to_bin(?)", action.ID.String())

	if action.ParentID != nil {
		row["parent_id"] = goqu.L("uuid_to_bin(?)", action.ParentID.String())
	} else {
		row["parent_id"] = nil
	}

	if action.ActorEntityID != nil && action.ActorEntityID.String() != "" {
		row["actor_entity_id"] = goqu.L("uuid_to_bin(?)", action.ActorEntityID.String())
	} else {
		row["actor_entity_id"] = nil
	}

	if action.TargetEntityID != nil && action.TargetEntityID.String() != "" {
		row["target_entity_id"] = goqu.L("uuid_to_bin(?)", action.TargetEntityID.String())
	} else {
		row["target_entity_id"] = nil
	}

	row["name"] = action.Name
	row["hash"] = action.Hash
	row["status"] = action.Status
	row["is_async"] = action.IsAsync
	row["emitted_at"] = action.EmittedAt.Time
	row["registered_at"] = action.RegisteredAt.Time

	if action.Details != nil {
		b, err := json.Marshal(action.Details)
		if err != nil {
			return "", nil, errors.Wrapf(err, "could not create details json string")
		}
		row["details"] = string(b)
	}

	if action.Delta != nil {
		b, err := json.Marshal(action.Delta)
		if err != nil {
			return "", nil, errors.Wrapf(err, "could not create delta json string")
		}
		row["delta"] = string(b)
	}

	return dialect.Insert("actions").Rows(row).ToSQL()
}
