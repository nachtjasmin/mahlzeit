-- name: AddDemoUser :one
insert into users(name, email, password_hash, password_hash_algorithm)
values ('demo user', 'demo@mahlzeit.app', '', 'argon2')
on conflict (email) do update set email = excluded.email
returning id;
