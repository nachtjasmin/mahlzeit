-- migrate:up
update steps
set time = null
where time < '0 minutes'::interval;

alter table steps
	add constraint "negative_time_check" check ( time >= '0 minutes'::interval);

-- migrate:down
alter table steps
	drop constraint negative_time_check;
