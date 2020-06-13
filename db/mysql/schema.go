package mysql

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type SQLMigrator struct {
	conn *sqlx.DB
}

func Migrator(conn *sqlx.DB) *SQLMigrator {
	return &SQLMigrator{
		conn: conn,
	}
}

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
		type VARCHAR(20),

		UNIQUE KEY unique_entity_and_name (entity_id, name),

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
	if _, err := m.conn.Exec(microserviceSchema); err != nil {
		return errors.Wrap(err, "could not create microservices table")
	}

	if _, err := m.conn.Exec(entitySchema); err != nil {
		return errors.Wrap(err, "could not create entities table")
	}

	if _, err := m.conn.Exec(eventSchema); err != nil {
		return errors.Wrap(err, "could not create events table")
	}

	if _, err := m.conn.Exec(propertySchema); err != nil {
		return errors.Wrap(err, "could not create property table")
	}

	if _, err := m.conn.Exec(changeSchema); err != nil {
		return errors.Wrap(err, "could not create change table")
	}

	return nil
}

func (m *SQLMigrator) Down() error {
	if _, err := m.conn.Exec(flush); err != nil {
		return errors.Wrap(err, "could not drop tables")
	}

	return nil
}
