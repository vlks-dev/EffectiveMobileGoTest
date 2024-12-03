CREATE TABLE "group" (
                       id SERIAL PRIMARY KEY,
                       group_name VARCHAR(255) NOT NULL,
                       song_name VARCHAR(255) NOT NULL,
                       release_date DATE NOT NULL,
                       text TEXT NOT NULL,
                       link VARCHAR(512) NOT NULL
);