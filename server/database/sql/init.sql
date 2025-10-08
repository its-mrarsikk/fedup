-- Table: feeds
CREATE TABLE IF NOT EXISTS feeds (
    id INTEGER PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
	link TEXT,
	fetchFrom TEXT,
    language TEXT,
    ttl INTEGER
);


-- Table: items
CREATE TABLE IF NOT EXISTS items (
    id INTEGER PRIMARY KEY,
    feed_id INTEGER NOT NULL,
    guid TEXT UNIQUE,
    title TEXT,
    description TEXT,
    link TEXT,
    author TEXT,
    pubDate TEXT,
    read BOOLEAN NOT NULL DEFAULT 0,
    enclosure_url TEXT,
    enclosure_type TEXT,
    enclosure_length INTEGER,
    FOREIGN KEY(feed_id) REFERENCES feeds(id)
);
