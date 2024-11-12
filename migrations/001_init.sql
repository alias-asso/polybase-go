CREATE TABLE IF NOT EXISTS courses (
	code TEXT,
	kind TEXT,
	part INTEGER DEFAULT 1,
	parts INTEGER DEFAULT 1,
	name TEXT,
	quantity INTEGER,
	total INTEGER,
	shown INTEGER DEFAULT 1,
	semester TEXT,
	PRIMARY KEY (code, kind, part)
);
