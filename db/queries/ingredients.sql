-- name: GetAllIngredients :many
select id, name
from ingredients
order by id;

-- name: GetIngredientNameByID :one
select name
from ingredients
where id = sqlc.arg('id');

-- name: AddIngredientToStep :exec
insert into step_ingredients (step_id, ingredients_id, unit_id, amount, note)
values (sqlc.arg('step_id'),
		sqlc.arg('ingredients_id'),
		nullif(sqlc.arg('unit_id')::bigint, 0),
		sqlc.arg('amount'),
		sqlc.arg('note'));

-- name: DeleteIngredientFromStep :exec
delete
from step_ingredients
where step_id = sqlc.arg('step_id')
  and ingredients_id = sqlc.arg('ingredients_id');

-- name: AddIngredient :one
insert into ingredients (name)
values (sqlc.arg('name'))
on conflict (name) do update set name=excluded.name -- no-op that effectively does nothing, but returns the ID as intended
returning id;
