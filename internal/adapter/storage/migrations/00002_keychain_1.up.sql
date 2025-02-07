BEGIN TRANSACTION;

CREATE TABLE
	keychain (
		id uuid PRIMARY KEY,
		"name" varchar,
		owner_id int8 NOT NULL,
		CONSTRAINT keychain_users_fk FOREIGN KEY (owner_id) REFERENCES users (id)
	);

CREATE TABLE
	keychain_item (
		id uuid PRIMARY KEY,
		keychain_id uuid NOT NULL,
		item_type int NOT NULL,
		label varchar,
		enc_key bytea,
		enc_value bytea,
		client_ts timestamp NOT NULL,
		server_ts timestamp NOT NULL,
		CONSTRAINT keychain_item_keychain_fk FOREIGN KEY (keychain_id) REFERENCES keychain (id)
	);

CREATE TABLE
	keychain_item_meta (
		keychain_id uuid NOT NULL,
		keychain_item_id uuid NOT NULL,
		k varchar,
		v varchar,
		PRIMARY key (keychain_id, keychain_item_id, k),
		CONSTRAINT keychain_fk FOREIGN KEY (keychain_id) REFERENCES keychain (id),
		CONSTRAINT keychain_item_fk FOREIGN KEY (keychain_item_id) REFERENCES keychain_item (id)
	);

COMMIT TRANSACTION;