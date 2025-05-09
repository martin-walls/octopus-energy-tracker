CREATE TABLE IF NOT EXISTS readings (
    timestamp TEXT PRIMARY KEY,
    total_consumption INTEGER NOT NULL,
    demand INTEGER NOT NULL
);
