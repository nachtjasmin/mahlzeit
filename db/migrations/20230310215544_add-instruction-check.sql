-- migrate:up
delete from steps where instruction = '';
alter table steps
	add constraint "instruction_length_check" check ( length(instruction) > 0 );

-- migrate:down
alter table steps
	drop constraint "instruction_length_check";
