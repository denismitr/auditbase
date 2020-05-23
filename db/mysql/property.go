package mysql

import (
	"database/sql"
	"github.com/denismitr/auditbase/model"
)

type property struct {
	ID          string         `db:"id"`
	Name        string         `db:"name"`
	EventID     string         `db:"event_id"`
	EntityID    string         `db:"entity_id"`
	ChangedFrom sql.NullString `db:"changed_from"`
	ChangedTo   sql.NullString `db:"changed_to"`
}

type propertyStat struct {
	Name       string `db:"name"`
	EventCount int    `db:"event_count"`
}

func (p *propertyStat) ToModel() *model.PropertyStat {
	return &model.PropertyStat{
		Name:       p.Name,
		EventCount: p.EventCount,
	}
}
