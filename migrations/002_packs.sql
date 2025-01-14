PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS packs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS pack_courses (
    pack_id INTEGER,
    course_code TEXT,
    course_kind TEXT,
    course_part INTEGER,
    FOREIGN KEY (pack_id) REFERENCES packs(id) ON DELETE CASCADE,
    FOREIGN KEY (course_code, course_kind, course_part) 
        REFERENCES courses(code, kind, part) ON UPDATE CASCADE,
    PRIMARY KEY (pack_id, course_code, course_kind, course_part)
);
