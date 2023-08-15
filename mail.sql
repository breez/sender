PRAGMA foreign_keys=OFF;
BEGIN TRANSACTION;
CREATE TABLE IF NOT EXISTS "configuration" (
	"name"	TEXT NOT NULL UNIQUE DEFAULT '',
	"value"	TEXT DEFAULT '',
	PRIMARY KEY("name")
) WITHOUT ROWID;
INSERT INTO configuration VALUES('Organization','');
INSERT INTO configuration VALUES('UID','');
INSERT INTO configuration VALUES('URL','');
INSERT INTO configuration VALUES('bodyHTML','');
INSERT INTO configuration VALUES('bodyText','');
INSERT INTO configuration VALUES('description','');
INSERT INTO configuration VALUES('end','');
INSERT INTO configuration VALUES('fromEmail','');
INSERT INTO configuration VALUES('fromName','');
INSERT INTO configuration VALUES('location','');
INSERT INTO configuration VALUES('organizerEmail','');
INSERT INTO configuration VALUES('organizerName','');
INSERT INTO configuration VALUES('smtpHost','');
INSERT INTO configuration VALUES('smtpPassword','');
INSERT INTO configuration VALUES('smtpPort','');
INSERT INTO configuration VALUES('smtpUsername','');
INSERT INTO configuration VALUES('start','');
INSERT INTO configuration VALUES('subject','');
INSERT INTO configuration VALUES('summary','');
CREATE TABLE IF NOT EXISTS "emails" (
	"email"	TEXT NOT NULL UNIQUE DEFAULT '',
	"first_name" TEXT NOT NULL DEFAULT '',
	"full_name" TEXT NOT NULL DEFAULT '',
	PRIMARY KEY("email")
) WITHOUT ROWID;
COMMIT;
