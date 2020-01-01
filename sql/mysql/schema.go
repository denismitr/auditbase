package mysql

const createMicroservices = `
CREATE TABLE IF NOT EXISTS microservices (
	id binary(16) PRIMARY KEY,
    name VARCHAR(36),
    description VARCHAR(255),
	created_at timestamp default current_timestamp,
	updated_at timestamp default current_timestamp on update current_timestamp,
	UNIQUE KEY unique_name (name)
);
`
