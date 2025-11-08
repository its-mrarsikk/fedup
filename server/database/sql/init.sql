CREATE TABLE IF NOT EXISTS feeds (
    id INTEGER PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
	link TEXT, -- HTML human-readable feed
	fetchFrom TEXT UNIQUE, -- RSS feed
    language TEXT,
    ttl INTEGER NOT NULL DEFAULT 60,
    etag TEXT,
    lastModified TEXT -- RFC 3339
);

CREATE TABLE IF NOT EXISTS items (
    id INTEGER PRIMARY KEY,
    feed_id INTEGER NOT NULL,
    guid TEXT UNIQUE,
    title TEXT,
    description TEXT,
    link TEXT,
    author TEXT,
    pubDate TEXT, -- RFC 3339
    read BOOLEAN NOT NULL DEFAULT 0,
    starred BOOLEAN NOT NULL DEFAULT 0,
    FOREIGN KEY(feed_id) REFERENCES feeds(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS enclosures (
    id INTEGER PRIMARY KEY,
    item_id INTEGER NOT NULL,
    type TEXT NOT NULL, -- MIME type of the enclosure
    url TEXT NOT NULL UNIQUE,
    -- length is not relevant in the modern age. it's not the size that matters, right?
    filePath TEXT NOT NULL,
    FOREIGN KEY(item_id) REFERENCES items(id) ON DELETE CASCADE
);

CREATE VIEW IF NOT EXISTS default_read AS SELECT * FROM items WHERE read = 0 OR starred = 1;
