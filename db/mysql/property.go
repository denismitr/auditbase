package mysql

import "database/sql"

type property struct {
	ID          string         `db:"id"`
	Name        string         `db:"name"`
	EventID     string         `db:"event_id"`
	EntityID    string         `db:"entity_id"`
	ChangedFrom sql.NullString `db:"changed_from"`
	ChangedTo   sql.NullString `db:"changed_to"`
}
