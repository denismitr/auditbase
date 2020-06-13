package mysql

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/denismitr/auditbase/model"
	"github.com/denismitr/auditbase/utils/logger"
	"github.com/denismitr/auditbase/utils/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type PropertyRepository struct {
	conn   *sqlx.DB
	log logger.Logger
	uuid4  uuid.UUID4Generator
}

func NewPropertyRepository(conn *sqlx.DB, uuid4 uuid.UUID4Generator, log logger.Logger) *PropertyRepository {
	return &PropertyRepository{
		conn: conn,
		uuid4: uuid4,
		log: log,
	}
}

type property struct {
	ID          string `db:"id"`
	Name        string `db:"name"`
	EntityID    string `db:"entity_id"`
	Type        string `db:"type"`
	ChangeCount int    `db:"change_count"`
}

func (p *property) ToModel() *model.Property {
	return &model.Property{
		ID:          p.ID,
		EntityID:    p.EntityID,
		Type:        p.Type,
		Name:        p.Name,
		ChangeCount: p.ChangeCount,
	}
}

func (r *PropertyRepository) GetIDOrCreate(name, entityID string) (string, error) {
	var result string

	_, err := sq.Insert("properties").
		Columns("id", "name", "entity_id").
		Values(
			sq.Expr("UUID_TO_BIN(?)", r.uuid4.Generate()),
			name,
			sq.Expr("UUID_TO_BIN(?)", entityID),
		).RunWith(r.conn).Query()

	if err != nil {
		r.log.Error(err)
	}

	id := sq.
		Select("BIN_TO_UUID(id) as id").
		From("properties").
		Where(sq.Eq{"name": name}).
		Where("entity_id = UUID_TO_BIN(?)", entityID).
		Limit(1)

	sql, args, _ := id.ToSql()
	r.log.Debugf("%s -- %#v", sql, args)

	rows, err := id.RunWith(r.conn).Query()
	if err != nil {
		return result, errors.Wrapf(err, "could not select id from property with name %s and entityID %s", name, entityID)
	}

	for rows.Next() {
		if err := rows.Scan(&result); err != nil {
			return result, errors.Wrapf(err, "could not parse property ID for name %s and entityID %s", name, entityID)
		}

		return result, nil
	}

	return result, errors.Errorf("failed to create or retrieve property with name %s and entityId %s", name, entityID)
}


