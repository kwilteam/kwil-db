CREATE TABLE users (
id int PRIMARY KEY,
name text CHECK(LENGTH(name) > 10),
address text NOT NULL DEFAULT 'usa',
email text NOT NULL UNIQUE ,
city_id int,
group_id int REFERENCES groups(id) ON UPDATE RESTRICT ON DELETE CASCADE,
CONSTRAINT city_fk FOREIGN KEY (city_id, address) REFERENCES cities(id, address) ON UPDATE NO ACTION ON DELETE SET NULL,
CHECK(LENGTH(email) > 1),
UNIQUE (city_id, address),
);