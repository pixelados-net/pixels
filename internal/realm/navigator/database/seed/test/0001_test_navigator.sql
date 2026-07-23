--liquibase formatted sql

--changeset pixels:pixels-navigator-seed-test-0001-test-navigator context:test
insert into room_favorites (player_id, room_id)
values
    (1, 1)
on conflict do nothing;

insert into navigator_preferences (player_id)
values
    (1)
on conflict do nothing;
--rollback delete from navigator_preferences where player_id = 1;
--rollback delete from room_favorites where player_id = 1;
