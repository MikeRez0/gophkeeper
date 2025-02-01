BEGIN TRANSACTION;

CREATE TABLE
	users (
		id bigserial PRIMARY KEY,
		login varchar NOT NULL,
		"password" varchar NOT NULL,
		CONSTRAINT users_unique UNIQUE (login)
	);

COMMIT TRANSACTION;