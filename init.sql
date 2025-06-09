create table if not exists authentication (id serial primary key, name text, email text, password text, type text, history text[]);
create table if not exists events (id serial primary key not null, name text, description text, invited text, start timestamp);
create table if not exists borrowed_books (id serial primary key not null, book_id int, user_id int, return_date date);
create table if not exists book_reservations (id serial primary key, book_id int, user_id int);
create table if not exists books (id serial primary key, isbn text, title text, author text, year int, quantity int);
create table if not exists reviews (id serial primary key, user_id int references authentication(id) on delete cascade, book_id int references books(id), stars numeric, comment text);
create table if not exists votes (id serial primary key, vote text, review_id int references reviews(id) on delete cascade, user_id int references authentication(id));
