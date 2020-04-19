package mysql

type property struct {
	ID          string `db:"id"`
	Name        string `db:"name"`
	EventID     string `db:"event_id"`
	ChangedFrom string `db:"changed_from"`
	ChangedTo   string `db:"changed_to"`
}
