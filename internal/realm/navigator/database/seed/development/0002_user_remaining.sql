--liquibase formatted sql

--changeset pixels:pixels-navigator-seed-development-0002-user-remaining context:development
insert into navigator_room_visits(player_id,room_id,visit_count,first_visited_at,last_visited_at)
values
    (1,2,8,now()-interval '7 days',now()-interval '1 hour'),
    (1,3,2,now()-interval '3 days',now()-interval '1 day'),
    (2,1,4,now()-interval '4 days',now()-interval '2 hours')
on conflict(player_id,room_id) do update set visit_count=excluded.visit_count,first_visited_at=excluded.first_visited_at,last_visited_at=excluded.last_visited_at;

insert into room_rights(room_id,player_id,granted_by_player_id)
values(2,1,2),(7,1,3)
on conflict do nothing;

update rooms set staff_picked=true where id=1;
--rollback update rooms set staff_picked=false where id=1;
--rollback delete from room_rights where player_id=1 and room_id in (2,7);
--rollback delete from navigator_room_visits where player_id in (1,2);
