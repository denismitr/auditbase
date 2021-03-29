package mysql

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/denismitr/auditbase/internal/db"
	"github.com/denismitr/auditbase/internal/model"
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
	ID             int            `db:"id"`
	ParentUID      sql.NullString `db:"parent_uid"`
	UID            string         `db:"uid"`
	Hash           string         `db:"hash"`
	ActorEntityID  sql.NullInt64  `db:"actor_entity_id"`
	TargetEntityID sql.NullInt64  `db:"target_entity_id"`
	Name           string         `db:"name"`
	Status         int8           `db:"status"`
	IsAsync        bool           `db:"is_async"`
	Details        sql.NullString `db:"details"`
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

	result, err := stmt.ExecContext(ctx, args...)
	if err != nil {
		return nil, errors.Wrap(err, "could not create action")
	}

	newID, err := result.LastInsertId()
	if err != nil {
		return nil, errors.Wrapf(err, "could not retrieve last insert ID on action [%s] create", action.Name)
	}

	return r.FirstByID(ctx, model.ID(newID))
}

func (r *ActionRepository) UpdateStatus(ctx context.Context, id model.ID, status model.Status) error {
	q, args, err := updateActionQuery(id, status)
	if err != nil {
		return errors.Wrap(err, "how could updateActionQuery func have failed?")
	}

	stmt, err := r.mysqlTx.PreparexContext(ctx, q)
	if err != nil {
		return errors.Wrapf(err, "could not prepare query %s", q)
	}

	if _, err := stmt.ExecContext(ctx, args...); err != nil {
		return  errors.Wrap(err, "could not update action status")
	}

	return nil
}

func updateActionQuery(id model.ID, status model.Status) (string, []interface{}, error) {
	q := "UPDATE actions SET status = ? WHERE id = ?"
	return q, []interface{}{int64(status), id.Int64()}, nil
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
	if err := selectStmt.SelectContext(ctx, &ars); err != nil {
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
		"id", "uid", "name",
		goqu.L("HEX(`hash`)").As("hash"),
		"parent_uid", "actor_entity_id", "target_entity_id",
		"is_async", "status",
		"emitted_at", "registered_at",
	).From("actions")

	if f.Has("uid") {
		exp := goqu.L("`uid` = ?", f.MustString("uid"))
		countQ = countQ.Where(exp)
		q = q.Where(exp)
	}

	if f.Has("parentUid") {
		exp := goqu.L("`parent_uid` = ?", f.MustString("parentUid"))
		countQ = countQ.Where(exp)
		q = q.Where(exp)
	}

	if f.Has("actorEntityId") {
		exp := goqu.L("`actor_entity_id` = ?", f.MustInt("actor_entity_id"))
		countQ = countQ.Where(exp)
		q = q.Where(exp)
	}

	if f.Has("targetEntityId") {
		exp := goqu.L("target_entity_id = ?", f.MustInt("target_entity_id"))
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

func (r *ActionRepository) FirstByUID(ctx context.Context, UID model.UID) (*model.Action, error) {
	q, args, err := firstActionByUIDQuery(UID)
	if err != nil {
		return nil, err
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
			return errors.Wrapf(db.ErrActionNotFound, "id [%d]", ID)
		}

		return errors.Wrapf(err, "could not delete action with ID [%d]", ID)
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
	return sq.Delete("actions").Where("`id` = ?", int(ID)).ToSql()
}

func firstActionByIDQuery(ID model.ID) (string, []interface{}, error) {
	dialect := goqu.Dialect(MySQL8)

	return dialect.From("actions").Select(
		"id", "uid", "parent_uid", "status", "is_async",
		goqu.L("HEX(`hash`)").As("hash"), "name", "details", "emitted_at", "registered_at",
	).Where(
		goqu.L("`id` = ?", int(ID)),
	).Limit(1).ToSQL()
}

func firstActionByUIDQuery(UID model.UID) (string, []interface{}, error) {
	dialect := goqu.Dialect(MySQL8)

	if !UID.Valid() {
		return "", nil, errors.Wrap(db.ErrInvalidQueryInput, "action uid is invalid")
	}

	return dialect.From("actions").Select(
		"id", "uid", "parent_uid", "is_async", "status",
		"actor_entity_id", "target_entity_id",
		goqu.L("HEX(`hash`)").As("hash"), "name", "details", "emitted_at", "registered_at",
	).Where(
		goqu.L("`uid` = ?", UID.String()),
	).Limit(1).ToSQL()
}

func createActionQuery(action *model.Action) (string, []interface{}, error) {
	dialect := goqu.Dialect(MySQL8)

	row := goqu.Record{}

	if action.ParentUID.Valid() {
		row["parent_uid"] = action.ParentUID.String()
	} else {
		row["parent_uid"] = nil
	}

	if action.ActorEntityID.Valid() {
		row["actor_entity_id"] = action.ActorEntityID.Int64()
	} else {
		row["actor_entity_id"] = nil
	}

	if action.TargetEntityID.Valid() {
		row["target_entity_id"] = action.TargetEntityID.Int64()
	} else {
		row["target_entity_id"] = nil
	}

	row["uid"] = action.UID.String()
	row["name"] = action.Name
	row["hash"] = goqu.L("UNHEX(?)", action.Hash)
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

	return dialect.Insert("actions").Rows(row).Prepared(true).ToSQL()
}
