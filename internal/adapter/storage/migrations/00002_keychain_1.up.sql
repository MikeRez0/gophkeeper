BEGIN TRANSACTION;

-- CREATE TYPE OrderStatus as enum ('NEW', 'PROCESSING', 'PROCESSED', 'INVALID');
CREATE TABLE
	keychain (
		id bigserial PRIMARY KEY,
		owner_id int8 NOT NULL,
		create_at timestamp NOT NULL,
		update_at timestamp NOT NULL,
		CONSTRAINT keychain_users_fk FOREIGN KEY (owner_id) REFERENCES users (id)
	);

CREATE TABLE
	keychain_item (
		id bigserial PRIMARY KEY,
		keychain_id int8 NOT NULL,
		item_type int NOT NULL,
		label varchar,
		enc_key bytea,
		enc_value bytea,
		create_at timestamp NOT NULL,
		update_at timestamp NOT NULL,
		CONSTRAINT keychain_item_keychain_fk FOREIGN KEY (keychain_id) REFERENCES keychain (id)
	);

CREATE TABLE
	keychain_item_meta (
		keychain_item_id int8,
		k varchar,
		v varchar,
		PRIMARY key (keychain_item_id, k),
		CONSTRAINT keychain_users_fk FOREIGN KEY (keychain_item_id) REFERENCES keychain_item (id)
	);

COMMIT TRANSACTION;