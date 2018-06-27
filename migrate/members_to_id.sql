do
$do$
declare
rec record;
people record;
begin
	for rec in (select room_id, members from rooms)
	loop
		for people in (select user_id, name from users)
		loop
			update rooms set members = array_replace(members, people.name, people.user_id) where room_id = rec.room_id;
		end loop;
	end loop;
end
$do$;
