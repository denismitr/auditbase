package mysql

import (
	"context"
	"database/sql"
	"github.com/denismitr/auditbase/utils/logger"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"sort"
)

type SQLMigrator struct {
	conn *sqlx.DB
	lg logger.Logger

	up map[string][]string
	applied map[string]bool
}

func Migrator(conn *sqlx.DB, lg logger.Logger) *SQLMigrator {
	m := &SQLMigrator{
		conn: conn,
		lg: lg,
		up: make(map[string][]string),
		applied: make(map[string]bool),
	}

	m.up["001_initial"] = []string{microserviceSchema, entitySchema, eventSchema, propertySchema, propertySchema, changeSchema}
	m.up["002_add_crud_to_events"] = []string{addCrudToEventSchema}

	return m
}

const migrationsSchema = `
CREATE TABLE IF NOT EXISTS migrations (
	name VARCHAR(36) PRIMARY KEY,
	created_at TIMESTAMP default CURRENT_TIMESTAMP
) ENGINE=INNODB;
`

const microserviceSchema = `
	CREATE TABLE IF NOT EXISTS microservices (
		id binary(16) PRIMARY KEY,
		name VARCHAR(36),
		description VARCHAR(255),
		created_at TIMESTAMP default CURRENT_TIMESTAMP,
		updated_at TIMESTAMP default CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		UNIQUE KEY unique_name (name)
	) ENGINE=INNODB;
`

const eventSchema = `
	CREATE TABLE IF NOT EXISTS events (
		id binary(16) PRIMARY KEY,
		parent_event_id binary(16),
		hash VARCHAR(40),
		actor_id VARCHAR(36) NOT NULL,
		actor_entity_id binary(16) NOT NULL,
		actor_service_id binary(16) NOT NULL,
		target_id VARCHAR(36) NOT NULL,
		target_entity_id binary(16),
		target_service_id binary(16) NOT NULL,
		event_name VARCHAR(36) NOT NULL,
		operation SMALLINT, 
		emitted_at TIMESTAMP NOT NULL,
		registered_at TIMESTAMP NOT NULL,

		FOREIGN KEY (actor_service_id)
        REFERENCES microservices(id)
		ON DELETE CASCADE,

		FOREIGN KEY (target_service_id)
        REFERENCES microservices(id)
		ON DELETE CASCADE,
		
		FOREIGN KEY (actor_entity_id)
        REFERENCES entities(id)
		ON DELETE CASCADE,
		
		FOREIGN KEY (target_entity_id)
        REFERENCES entities(id)
        ON DELETE CASCADE
	) ENGINE=INNODB;
`

const entitySchema = `
	CREATE TABLE IF NOT EXISTS entities (
		id binary(16) PRIMARY KEY,
		service_id binary(16) NOT NULL,
		name VARCHAR(64) NOT NULL,
		description VARCHAR(255),
		created_at TIMESTAMP default CURRENT_TIMESTAMP,
		updated_at TIMESTAMP default CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		UNIQUE KEY unique_service_and_name (service_id, name),

		FOREIGN KEY (service_id)
        REFERENCES microservices(id)
        ON DELETE CASCADE
	) ENGINE=INNODB;		
`

const propertySchema = `
	CREATE TABLE IF NOT EXISTS properties (
		id binary(16) PRIMARY KEY,
		entity_id binary(16) NOT NULL,
		name VARCHAR(64) NOT NULL,
		
		UNIQUE KEY unique_entity_and_name (entity_id, name),
		INDEX name_index (name),

		FOREIGN KEY (entity_id)
        REFERENCES entities(id)
        ON DELETE CASCADE
	)
`

const changeSchema = `
	CREATE TABLE IF NOT EXISTS changes (
		id binary(16) PRIMARY KEY,
		property_id binary(16) NOT NULL,
		event_id binary(16) NOT NULL,
		from_value TEXT,
		to_value TEXT,
		current_data_type TINYINT(1),

		INDEX event_and_property_index (event_id, property_id),
		INDEX event_index (event_id),

		FOREIGN KEY (event_id)
        REFERENCES events(id)
        ON DELETE CASCADE,

		FOREIGN KEY (property_id)
        REFERENCES properties(id)
        ON DELETE CASCADE
	)
`

const addCrudToEventSchema = `
	ALTER TABLE events ADD COLUMN crud TINYINT(1) AFTER event_name 
`



const flush = `
	SET FOREIGN_KEY_CHECKS=0;

	DROP TABLE IF EXISTS properties;
	DROP TABLE IF EXISTS changes;
	DROP TABLE IF EXISTS entities;
	DROP TABLE IF EXISTS microservices;
	DROP TABLE IF EXISTS events; 

	SET FOREIGN_KEY_CHECKS=1;
`

func (m *SQLMigrator) Up() error {
	m.lg.Debugf("Acquiring exclusive lock for the whole DB...")
	if _, err := m.conn.Exec("SELECT GET_LOCK('migrations', 10)"); err != nil {
		return errors.Wrap(err, "could not obtain 'migrations' exclusive DB lock")
	}

	if _, err := m.conn.Exec(migrationsSchema); err != nil {
		return err
	}

	tx, err := m.conn.BeginTxx(context.Background(), &sql.TxOptions{
		Isolation: sql.LevelSerializable,
		ReadOnly: false,
	})

	if err != nil {
		return err
	}

	rows, err := tx.Queryx("SELECT name FROM migrations")
	if err != nil {
		_ = tx.Rollback()
		m.conn.Exec("SELECT RELEASE_LOCK('migrations')")
		return err
	}

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			_ = tx.Rollback()
			m.conn.Exec("SELECT RELEASE_LOCK('migrations')")
			return err
		}
		m.applied[name] = true
	}

	migrations := make([]string, 0, len(m.up))
	for name := range m.up {
		migrations = append(migrations, name)
	}

	sort.Strings(migrations)

	for _, name := range migrations {
		if _, ok := m.applied[name]; !ok {
			m.lg.Debugf("Running migration %s...", name)

			for _, query := range m.up[name] {
				m.lg.Debugf("Running SQL %s...", query)
				if _, err := tx.Exec(query); err != nil {
					_ = tx.Rollback()
					m.conn.Exec("SELECT RELEASE_LOCK('migrations')")
					return errors.Wrapf(err, "could not apply migration %s", name)
				}
			}

			if _, err := tx.Exec("INSERT INTO migrations (name) VALUES (?)", name); err != nil {
				_ = tx.Rollback()
				m.conn.Exec("SELECT RELEASE_LOCK('migrations')")
				return err
			}
		} else {
			m.lg.Debugf("Migration %s already applied", name)
		}
	}

	if err := tx.Commit(); err != nil {
		m.conn.Exec("SELECT RELEASE_LOCK('migrations')")
		return err
	}

	if _, err := m.conn.Exec("SELECT RELEASE_LOCK('migrations')"); err != nil {
		return err
	}

	m.lg.Debugf("Migrations are finished...")
	m.lg.Debugf("All locks are released...")

	return nil
}

func (m *SQLMigrator) Down() error {
	if _, err := m.conn.Exec(flush); err != nil {
		return errors.Wrap(err, "could not drop tables")
	}

	return nil
}
