-- name: GetAllIngredients :many
select name
from ingredients
order by id;
