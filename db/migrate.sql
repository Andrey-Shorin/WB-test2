create table city (id SERIAL  primary key , name text UNIQUE,country text, lat  FLOAT  NOT NULL, lon FLOAT  NOT NULL);
create table  forecast (id SERIAL  primary key , city text, temp FLOAT,time timestamp, data text);
ALTER TABLE forecast ADD CONSTRAINT unique_city_time UNIQUE (city, time);