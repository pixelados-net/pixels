--liquibase formatted sql

--changeset pixels:furniture-seed-football-gate-kit-0043 context:development
update furniture_definitions set
 custom_params='ch-255-66.lg-275-66.sh-295-66',
 updated_at=now()
where id=950023;

update furniture_items set
 x=10,y=11,updated_at=now(),version=version+1
where id=961405 and room_id=152;

--rollback update furniture_definitions set custom_params='ch-255-66.ca-1808-66.cc-3039-66.cp-3035-66.lg-275-66.wa-2001-66.sh-295-66',updated_at=now() where id=950023; update furniture_items set x=9,y=11,updated_at=now(),version=version+1 where id=961405 and room_id=152;
