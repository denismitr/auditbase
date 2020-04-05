package mysql

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

const microservicesSchema = `
	CREATE TABLE IF NOT EXISTS microservices (
		id binary(16) PRIMARY KEY,
		name VARCHAR(36),
		description VARCHAR(255),
		created_at TIMESTAMP default CURRENT_TIMESTAMP,
		updated_at TIMESTAMP default CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		UNIQUE KEY unique_name (name),
		INDEX name_index (name)
	) ENGINE=INNODB;
`

const eventsSchema = `
	CREATE TABLE IF NOT EXISTS events (
		id binary(16) PRIMARY KEY,
		parent_event_id binary(16) DEFAULT NULL,
		actor_id VARCHAR(36) NOT NULL,
		actor_type_id binary(16) NOT NULL,
		actor_service_id binary(16) NOT NULL,
		target_id VARCHAR(36) NOT NULL,
		target_type_id binary(16),
		target_service_id binary(16) NOT NULL,
		event_name VARCHAR(36) NOT NULL,
		emitted_at TIMESTAMP NOT NULL,
		registered_at TIMESTAMP NOT NULL,
		delta JSON DEFAULT NULL,

		FOREIGN KEY (actor_service_id)
        REFERENCES microservices(id)
		ON DELETE CASCADE,

		FOREIGN KEY (target_service_id)
        REFERENCES microservices(id)
		ON DELETE CASCADE,
		
		FOREIGN KEY (actor_type_id)
        REFERENCES actor_types(id)
		ON DELETE CASCADE,
		
		FOREIGN KEY (target_type_id)
        REFERENCES target_types(id)
        ON DELETE CASCADE
	) ENGINE=INNODB;
`

const actorTypeSchema = `
	CREATE TABLE IF NOT EXISTS actor_types (
		id binary(16) PRIMARY KEY,
		name VARCHAR(36),
		description VARCHAR(255),
		created_at TIMESTAMP default CURRENT_TIMESTAMP,
		updated_at TIMESTAMP default CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		UNIQUE KEY unique_name (name),
		INDEX name_index (name)
	) ENGINE=INNODB;
`

const targetTypeSchema = `
	CREATE TABLE IF NOT EXISTS target_types (
		id binary(16) PRIMARY KEY,
		name VARCHAR(36),
		description VARCHAR(255),
		created_at TIMESTAMP default CURRENT_TIMESTAMP,
		updated_at TIMESTAMP default CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		UNIQUE KEY unique_name (name),
		INDEX name_index (name)
	) ENGINE=INNODB;		
`

const flush = `
	SET SQL_SAFE_UPDATES = 0;

	truncate table events;
	delete from actor_types;
	delete from target_types;
	delete from microservices;

	SET SQL_SAFE_UPDATES = 1;
`

func Scaffold(conn *sqlx.DB) error {
	if _, err := conn.Exec(targetTypeSchema); err != nil {
		return errors.Wrap(err, "could not create target_types table")
	}

	if _, err := conn.Exec(actorTypeSchema); err != nil {
		return errors.Wrap(err, "could not create actor_types table")
	}

	if _, err := conn.Exec(microservicesSchema); err != nil {
		return errors.Wrap(err, "could not create microservices table")
	}

	if _, err := conn.Exec(eventsSchema); err != nil {
		return errors.Wrap(err, "could not create events table")
	}

	return nil
}

func Drop(conn *sqlx.DB) error {
	drop := `
		DROP TABLE events;
		DROP TABLE microservices;
		DROP TABLE target_types;
		DROP TABLE actor_types;
	`

	if _, err := conn.Exec(drop); err != nil {
		return errors.Wrap(err, "could not drop tables")
	}

	return nil
}
