package mysql

import (
	"database/sql"
	"github.com/denismitr/auditbase/model"
)

type change struct {
	ID         string         `db:"id"`
	EventID    string         `db:"event_id"`
	PropertyID string         `db:"property_id"`
	FromValue  sql.NullString `db:"from_value"`
	ToValue    sql.NullString `db:"to_value"`
}

type propertyChange struct {
	ID           string         `db:"id"`
	EventID      string         `db:"event_id"`
	PropertyID   string         `db:"property_id"`
	FromValue    sql.NullString `db:"from_value"`
	ToValue      sql.NullString `db:"to_value"`
	PropertyName string         `db:"property_name"`
	EntityID     string         `db:"entity_id"`
}

func (c *propertyChange) ToModel() *model.PropertyChange {
	return &model.PropertyChange{
		ID:           c.ID,
		EventID:      c.EventID,
		EntityID:     c.EntityID,
		From:         &c.FromValue.String,
		To:           &c.ToValue.String,
		PropertyID:   c.PropertyID,
		PropertyName: c.PropertyName,
	}
}

func (c *change) ToModel() *model.Change {
	return &model.Change{
		ID:         c.ID,
		EventID:    c.EventID,
		From:       &c.FromValue.String,
		To:         &c.ToValue.String,
		PropertyID: c.PropertyID,
	}
}
