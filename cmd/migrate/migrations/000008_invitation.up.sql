CREATE TABLE IF NOT EXISTS invitations (
	token bytea NOT NULL,
	user_id bigint NOT NULL,
	expiry timestamp NOT NULL,
	PRIMARY KEY (token, user_id)
);
