--liquibase formatted sql

--changeset pixels:pixels-room-0013-social-group-binding
alter table room_social_groups
    add column created_at timestamptz not null default now(),
    add column updated_at timestamptz not null default now(),
    add column version bigint not null default 1,
    add constraint room_social_groups_version_chk check(version > 0);

create unique index room_social_groups_group_uidx on room_social_groups(group_id);

update social_groups groups
set home_room_id=bindings.room_id
from room_social_groups bindings
where groups.id=bindings.group_id and groups.home_room_id is null;

alter table social_groups
    add constraint social_groups_home_room_fk foreign key(home_room_id) references rooms(id),
    add constraint social_groups_active_home_room_chk check(deactivated_at is not null or home_room_id is not null);

--rollback alter table social_groups drop constraint if exists social_groups_active_home_room_chk; alter table social_groups drop constraint if exists social_groups_home_room_fk; drop index if exists room_social_groups_group_uidx; alter table room_social_groups drop column version,drop column updated_at,drop column created_at;
