CREATE TABLE version (
    id INTEGER PRIMARY KEY,
    version VARCHAR(255)
);

CREATE TABLE user (
    id INTEGER PRIMARY KEY,
    username VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) DEFAULT NULL,
    is_admin BIT DEFAULT 0
);

CREATE TABLE logentry (
    id INTEGER PRIMARY KEY,
    user_id INTEGER NOT NULL,
    time DATETIME NOT NULL,
    food VARCHAR(255),
    notes VARCHAR(255),
    FOREIGN KEY (user_id) REFERENCES user(id)
)