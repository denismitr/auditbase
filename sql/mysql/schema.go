package mysql

const microservicesSchema = `
	CREATE TABLE IF NOT EXISTS microservices (
		id binary(16) PRIMARY KEY,
		name VARCHAR(36),
		description VARCHAR(255),
		created_at TIMESTAMP default CURRENT_TIMESTAMP,
		updated_at TIMESTAMP default CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		UNIQUE KEY unique_name (name)
	);
`

const eventsSchema = `
	CREATE TABLE IF NOT EXISTS events (
		id binary(16) PRIMARY KEY,
		parent_event_id binary(16) DEFAULT NULL,
		actor_id VARCHAR(36) NOT NULL,
		actor_type VARCHAR(36) NOT NULL,
		actor_service_id binary(16),
		target_id VARCHAR(36) NOT NULL,
		target_type VARCHAR(36),
		target_service_id binary(16) NOT NULL,
		event_name VARCHAR(36) NOT NULL,
		emitted_at TIMESTAMP NOT NULL,
		registered_at TIMESTAMP NOT NULL,
		delta JSON DEFAULT NULL
	);
`
