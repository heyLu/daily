CREATE TABLE IF NOT EXISTS entries (
	`id`    VARCHAR(16) PRIMARY KEY,
	`date`  TIMESTAMP NOT NULL,
	`type`  TEXT NOT NULL,
	`note`  TEXT NOT NULL,
	`value` FLOAT,
	`data`  TEXT
);
