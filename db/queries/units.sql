-- name: GetAllUnits :many
select id, name
from units
order by name;
