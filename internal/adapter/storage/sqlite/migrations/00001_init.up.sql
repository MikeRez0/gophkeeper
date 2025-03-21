CREATE TABLE
	users (
		id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL UNIQUE,
		login varchar NOT NULL,
		"password" varchar NOT NULL,
		CONSTRAINT users_unique UNIQUE (login)
	);
