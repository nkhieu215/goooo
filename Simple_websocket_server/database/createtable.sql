-- tao user_info
CREATE TABLE IF NOT EXISTS user_info (
	user_id serial PRIMARY KEY,
	name VARCHAR UNIQUE NOT NULL,
	password VARCHAR ,
	email VARCHAR UNIQUE ,
    token VARCHAR ,
	created_on TIMESTAMP NOT NULL,
        last_login TIMESTAMP 
);

-- create room_info
CREATE TABLE IF NOT EXISTS room_info(
    room_id serial PRIMARY KEY,
    room_name VARCHAR UNIQUE NOT NULL,
    time_created TIMESTAMP NOT NULL,
    creater INT NOT NULL
);

-- creater members
CREATE TABLE IF NOT EXISTS members(
    stt serial PRIMARY KEY,
    room_id INT NOT NULL,
    user_id INT NOT NULL,
    time_apply TIMESTAMP NOT NULL
);

-- create messages
CREATE TABLE IF NOT EXISTS messages(
    id serial PRIMARY KEY,
    time TIMESTAMP NOT NULL,
    sender_id INT NOT NULL,
    receiver_id INT ,
    room_id INT NOT NULL,
    message VARCHAR 
);
