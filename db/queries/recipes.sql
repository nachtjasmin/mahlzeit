-- name: GetAllRecipesByName :many
select id, name
from recipes
order by name;

-- name: GetRecipeByID :one
select id,
	   name,
	   description,
	   working_time,
	   waiting_time,
	   created_at,
	   updated_at,
	   created_by,
	   source,
	   servings,
	   servings_description
from recipes
where id = sqlc.arg(id);

-- name: GetStepsForRecipeByID :many
select steps.id,
	   instruction,
	   "time"  as step_time,
	   -- To get the ingredients within the same query and avoiding n+1 query pipelines,
	   -- those are built as a JSON object using jsonb_build_object.
	   -- Because the values can have NULL values due to the left join below, we strip those values
	   -- with jsonb_strip_nulls. And in the end, they are grouped inside an array with jsonb_agg.
	   jsonb_agg(jsonb_strip_nulls(jsonb_build_object(
			   'name', ingredients.name,
			   'amount', step_ingredients.amount,
			   'note', step_ingredients.note
		   ))) as ingredients
from steps
		 left join step_ingredients on steps.id = step_ingredients.step_id
		 left join ingredients on step_ingredients.ingredients_id = ingredients.id
where steps.recipe_id = sqlc.arg(id)
group by steps.id, "time", instruction;


-- name: GetTotalIngredientsForRecipe :many
select ingredients.name,
	   sum(step_ingredients.amount) as total_amount
from steps
		 inner join step_ingredients on steps.id = step_ingredients.step_id
		 inner join ingredients on ingredients.id = step_ingredients.ingredients_id
where steps.recipe_id = sqlc.arg(id)
group by ingredients.name
order by ingredients.name, total_amount desc;

-- name: UpdateBasicRecipeInformation :exec
update recipes
set name        = sqlc.arg('name'),
	servings    = sqlc.arg('servings'),
	description = sqlc.arg('description'),
	updated_at  = now()
where id = sqlc.arg('id');

-- name: DeleteStepByID :exec
delete
from steps
where id = sqlc.arg('id');

-- name: UpdateStepByID :exec
update steps
set instruction = sqlc.arg('instruction'),
	time        = sqlc.arg('time')
where id = sqlc.arg('id');
