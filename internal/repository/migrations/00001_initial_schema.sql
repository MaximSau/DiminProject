-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users (
	id TEXT PRIMARY KEY,
	code TEXT UNIQUE NOT NULL,
	created_at DATETIME NOT NULL,
	updated_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS profiles (
	user_id TEXT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
	name TEXT,
	age INTEGER,
	height_cm REAL,
	weight_kg REAL,
	sex TEXT,
	goal TEXT,
	pace TEXT,
	level TEXT,
	workouts_per_week INTEGER,
	session_minutes INTEGER,
	equipment TEXT,
	preferred_time TEXT,
	activity_level TEXT,
	diet_preferences TEXT,
	allergies TEXT,
	avoid_foods TEXT,
	health_flags TEXT,
	health_notes TEXT,
	consent INTEGER,
	onboarded INTEGER,
	plan_version INTEGER
);

CREATE TABLE IF NOT EXISTS plans (
	user_id TEXT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
	plan_json TEXT
);

CREATE TABLE IF NOT EXISTS reminders (
	id TEXT NOT NULL,
	user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	kind TEXT,
	title TEXT,
	time TEXT,
	days TEXT,
	enabled INTEGER,
	note TEXT,
	PRIMARY KEY(user_id, id)
);

CREATE TABLE IF NOT EXISTS weights (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	date TEXT NOT NULL,
	kg REAL NOT NULL,
	note TEXT
);

CREATE TABLE IF NOT EXISTS sessions (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	date TEXT NOT NULL,
	day_index INTEGER NOT NULL,
	title TEXT NOT NULL,
	completed INTEGER NOT NULL,
	feeling TEXT,
	energy INTEGER
);

CREATE TABLE IF NOT EXISTS checkins (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	date TEXT NOT NULL,
	mood INTEGER,
	energy INTEGER,
	nutrition INTEGER,
	sleep REAL,
	note TEXT
);

CREATE TABLE IF NOT EXISTS eaten_meals (
	user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	meal_key TEXT NOT NULL,
	PRIMARY KEY(user_id, meal_key)
);

CREATE TABLE IF NOT EXISTS water_logs (
	user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	date TEXT NOT NULL,
	ml INTEGER NOT NULL,
	PRIMARY KEY(user_id, date)
);

CREATE TABLE IF NOT EXISTS chat_messages (
	id TEXT NOT NULL,
	user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	role TEXT NOT NULL,
	text TEXT NOT NULL,
	at DATETIME NOT NULL,
	kind TEXT,
	suggestions TEXT,
	PRIMARY KEY(user_id, id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS chat_messages;
DROP TABLE IF EXISTS water_logs;
DROP TABLE IF EXISTS eaten_meals;
DROP TABLE IF EXISTS checkins;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS weights;
DROP TABLE IF EXISTS reminders;
DROP TABLE IF EXISTS plans;
DROP TABLE IF EXISTS profiles;
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
