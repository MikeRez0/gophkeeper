CREATE USER gophkeeper
    PASSWORD 'gophkeeper';

CREATE DATABASE gophkeeper_db
    OWNER 'gophkeeper'
    ENCODING 'UTF8'
    LC_COLLATE = 'en_US.utf8'
    LC_CTYPE = 'en_US.utf8';