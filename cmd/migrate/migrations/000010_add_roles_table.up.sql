CREATE TABLE IF NOT EXISTS roles (
	id BIGSERIAL PRIMARY KEY,
	name VARCHAR(255) NOT NULL UNIQUE,
	level int NOT NULL DEFAULT 0,
	description TEXT
);

INSERT INTO roles (name, level, description) VALUES
	('user', 1, 'Regular user'),
	('moderator', 2, 'Moderator'),
	('admin', 3, 'Administrator');
