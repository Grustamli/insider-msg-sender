CREATE TABLE IF NOT EXISTS message
(
    id         SERIAL PRIMARY KEY,
    recipient  VARCHAR NOT NULL,
    content    TEXT    NOT NULL,
    message_id VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    sent_at    TIMESTAMP

);