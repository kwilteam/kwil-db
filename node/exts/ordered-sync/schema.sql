-- this file implements the schema for syncing data from any chain
SET CURRENT NAMESPACE TO kwil_ordered_sync;

CREATE TABLE topics (
    id UUID PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    resolve_func TEXT NOT NULL,
    last_processed_point int8
);

CREATE TABLE pending_data (
    point int8,
    topic_id UUID REFERENCES topics(id) ON UPDATE CASCADE ON DELETE CASCADE,
    previous_point int8, -- can be null if this is the first point
    data bytea not null,
    PRIMARY KEY (point, topic_id)
    -- we dont include an fk for previous point because it
    -- can be null if events are processed out of order (which is
    -- what this package is designed to handle)
);

CREATE TABLE meta(
    version int8 PRIMARY KEY
)