PRAGMA foreign_keys = ON;

CREATE TABLE feeds (
    id INTEGER PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    link TEXT NOT NULL,
    fetch_from TEXT NOT NULL UNIQUE,
    language TEXT,
    ttl INTEGER DEFAULT 60 NOT NULL,
    etag TEXT,
    lastModified DATETIME
);

CREATE TABLE items (
    id INTEGER PRIMARY KEY,
    feedId INTEGER NOT NULL,
    guid TEXT NOT NULL UNIQUE,
    title TEXT,
    description TEXT,
    link TEXT,
    author TEXT,
    pubDate DATETIME,
    read INTEGER DEFAULT 0,
    starred INTEGER DEFAULT 0,

    FOREIGN KEY(`feedId`) REFERENCES feeds(`id`) ON DELETE CASCADE
);

CREATE TABLE enclosures (
    id INTEGER PRIMARY KEY,
    itemId INTEGER NOT NULL,
    mimeType TEXT NOT NULL,
    url TEXT NOT NULL,

    FOREIGN KEY(`itemId`) REFERENCES items(`id`) ON DELETE CASCADE
);

CREATE VIEW default_items AS
 SELECT items.* WHERE items.read = 0 OR items.starred = 1;
