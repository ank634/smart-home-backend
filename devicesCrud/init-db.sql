create table IF NOT EXISTS Room(
	id Serial Primary KEY,
	name TEXT NOT NULL UNIQUE
	CHECK(TRIM(name) <> '')
);

create type device_type as ENUM ('light');
create type manufactor_type as ENUM ('custom');
create type service_type as ENUM ('http._tcp');

create table IF NOT EXISTS Device(
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL UNIQUE,
	serviceType service_type NOT NULL,
	deviceType device_type NOT NULL,
	manufactor manufactor_type NOT NULL,
	setTopic TEXT NOT NULL UNIQUE,
	getTopic TEXT NOT NULL UNIQUE,
	endpoint TEXT NOT NULL UNIQUE,
	room int,
	FOREIGN KEY (room) REFERENCES ROOM(id) ON DELETE SET NULL,
	
	CHECK(TRIM(id) <> ''),
	CHECK(TRIM(name) <> ''),
	CHECK(TRIM(setTopic) <> ''),
	CHECK(TRIM(getTopic) <> ''),
	CHECK(TRIM(endpoint) <> ''),

	CHECK(TRIM(name) = name),
	CHECK(TRIM(setTopic) = setTopic),
	CHECK(TRIM(getTopic) = getTopic),
	CHECK(TRIM(endpoint) = endpoint)
);

create table IF NOT EXISTS light(
	id TEXT Primary KEY,
	dimmable boolean NOT NULL,
	rgb boolean NOT NULL,
	FOREIGN KEY (id) REFERENCES Device(id) ON DELETE CASCADE,
	CHECK(TRIM(id) <> '')
);