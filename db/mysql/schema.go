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

	m.up["001_initial"] = []string{microservicesSchema, entityTypesSchema, entitiesSchema, actionsSchema}

	return m
}

const migrationsSchema = `
CREATE TABLE IF NOT EXISTS migrations (
	name VARCHAR(36) PRIMARY KEY,
	created_at TIMESTAMP default CURRENT_TIMESTAMP
) ENGINE=INNODB;
`

const microservicesSchema = `
	CREATE TABLE IF NOT EXISTS microservices (
		id BIGINT UNSIGNED AUTO_INCREMENT,
		name VARCHAR(36),
		description VARCHAR(255),
		created_at TIMESTAMP default CURRENT_TIMESTAMP,
		updated_at TIMESTAMP default CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		UNIQUE KEY unique_name (name),

		PRIMARY KEY (id)
	) ENGINE=INNODB;
`

const actionsSchema = `
	CREATE TABLE IF NOT EXISTS actions (
		id BIGINT UNSIGNED AUTO_INCREMENT,
		parent_id BIGINT UNSIGNED,
		status TINYINT(1) DEFAULT 0,
		is_async TINYINT(1) DEFAULT 0,
		hash VARCHAR(40),
		actor_entity_id BIGINT UNSIGNED,
		target_entity_id BIGINT UNSIGNED,
		name VARCHAR(36) NOT NULL,
		emitted_at TIMESTAMP NOT NULL,
		registered_at TIMESTAMP NOT NULL,
		details JSON,

		PRIMARY KEY (id),

		INDEX name_idx (name),

		FOREIGN KEY (actor_entity_id)
        REFERENCES entities(id)
		ON DELETE CASCADE,

		FOREIGN KEY (target_entity_id)
        REFERENCES entities(id)
		ON DELETE CASCADE
	) ENGINE=INNODB;
`

const entityTypesSchema = `
	CREATE TABLE IF NOT EXISTS entity_types (
		id BIGINT UNSIGNED AUTO_INCREMENT,
		service_id BIGINT UNSIGNED NOT NULL,
		name VARCHAR(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
		description VARCHAR(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
		is_actor TINYINT (1) NOT NULL DEFAULT 0, 
		created_at TIMESTAMP default CURRENT_TIMESTAMP,
		updated_at TIMESTAMP default CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

		PRIMARY KEY (id),

		UNIQUE KEY unique_service_and_name (service_id, name),

		FOREIGN KEY (service_id)
        REFERENCES microservices(id)
        ON DELETE CASCADE
	) ENGINE=INNODB;		
`

const entitiesSchema = `
	CREATE TABLE IF NOT EXISTS entities (
		id BIGINT UNSIGNED AUTO_INCREMENT,
		entity_type_id BIGINT UNSIGNED NOT NULL,
		external_id VARCHAR(36) NOT NULL,
		created_at TIMESTAMP default CURRENT_TIMESTAMP,
		updated_at TIMESTAMP default CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

		UNIQUE KEY unique_idx (entity_type_id, external_id),

		PRIMARY KEY (id)
	) ENGINE=INNODB;
`

const flush = `
	SET FOREIGN_KEY_CHECKS=0;

	DROP TABLE IF EXISTS microservices;
	DROP TABLE IF EXISTS actions; 

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
		_, _ = m.conn.Exec("SELECT RELEASE_LOCK('migrations')")
		return err
	}

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			_ = tx.Rollback()
			_, _ = m.conn.Exec("SELECT RELEASE_LOCK('migrations')")
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
					_, _ = m.conn.Exec("SELECT RELEASE_LOCK('migrations')")
					return errors.Wrapf(err, "could not apply migration %s", name)
				}
			}

			if _, err := tx.Exec("INSERT INTO migrations (name) VALUES (?)", name); err != nil {
				_ = tx.Rollback()
				_, _ = m.conn.Exec("SELECT RELEASE_LOCK('migrations')")
				return err
			}
		} else {
			m.lg.Debugf("Migration %s already applied", name)
		}
	}

	if err := tx.Commit(); err != nil {
		_, _ = m.conn.Exec("SELECT RELEASE_LOCK('migrations')")
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
