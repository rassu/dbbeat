create table if not exists person (id smallserial primary key, name text, age int, gender char(1));
create table if not exists car (id smallserial primary key, manufacturer text, year int, mileage decimal, price decimal, isNew bool);
create table if not exists sample (id serial primary key, name char(3), arr integer[], nestedarr integer[][]);
