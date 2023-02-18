-- name: GetAllRecipesByName :many
select id, name
from recipes
order by name;
