--liquibase formatted sql

--changeset pixels:pixels-camera-0006-capture-lifecycle
alter table camera_captures
    add column state text not null default 'pending',
    add column superseded_at timestamptz null,
    add column abandoned_at timestamptz null,
    add column deleted_at timestamptz null,
    add column cleanup_attempted_at timestamptz null,
    add column purchase_count integer not null default 0 check (purchase_count >= 0),
    add constraint camera_captures_state_check check (state in ('pending','purchased','published','purchased_published','superseded','abandoned','deleted'));

create table camera_capture_items (
    capture_id bigint not null references camera_captures(id) on delete cascade,
    item_id bigint not null unique references furniture_items(id) on delete cascade,
    created_at timestamptz not null default now(),
    primary key (capture_id, item_id)
);

insert into camera_capture_items (capture_id, item_id)
select capture.id, item.id
from camera_captures capture
join furniture_items item on item.definition_id=940001
    and item.extra_data like '%"u"%'
    and item.extra_data::jsonb->>'u'=capture.capture_uuid::text
on conflict do nothing;

update camera_captures capture
set purchase_count=(select count(*) from camera_capture_items link where link.capture_id=capture.id),
    state=case
        when capture.kind='photo'
             and capture.id=(select latest.id from camera_captures latest where latest.player_id=capture.player_id and latest.kind='photo' order by latest.created_at desc,latest.id desc limit 1)
             and exists(select 1 from camera_capture_items link where link.capture_id=capture.id)
             and exists(select 1 from camera_publications publication where publication.capture_id=capture.id and publication.removed_at is null)
            then 'purchased_published'
        when capture.kind='photo'
             and capture.id=(select latest.id from camera_captures latest where latest.player_id=capture.player_id and latest.kind='photo' order by latest.created_at desc,latest.id desc limit 1)
             and exists(select 1 from camera_publications publication where publication.capture_id=capture.id and publication.removed_at is null)
            then 'published'
        when capture.kind='photo'
             and capture.id=(select latest.id from camera_captures latest where latest.player_id=capture.player_id and latest.kind='photo' order by latest.created_at desc,latest.id desc limit 1)
             and exists(select 1 from camera_capture_items link where link.capture_id=capture.id)
            then 'purchased'
        when capture.kind='photo'
             and capture.id<>(select latest.id from camera_captures latest where latest.player_id=capture.player_id and latest.kind='photo' order by latest.created_at desc,latest.id desc limit 1)
            then 'superseded'
        else 'pending'
    end,
    superseded_at=case when capture.kind='photo'
        and capture.id<>(select latest.id from camera_captures latest where latest.player_id=capture.player_id and latest.kind='photo' order by latest.created_at desc,latest.id desc limit 1)
        then coalesce(capture.consumed_at,capture.created_at) else null end;

drop index camera_captures_one_pending_idx;
create unique index camera_captures_one_active_idx on camera_captures(player_id)
where kind='photo' and state in ('pending','purchased','published','purchased_published');
create index camera_captures_cleanup_idx on camera_captures(state, created_at)
where kind='photo' and state in ('pending','superseded','abandoned');
create index camera_capture_items_capture_idx on camera_capture_items(capture_id);

--rollback drop index if exists camera_capture_items_capture_idx;
--rollback drop index if exists camera_captures_cleanup_idx;
--rollback drop index if exists camera_captures_one_active_idx;
--rollback create unique index camera_captures_one_pending_idx on camera_captures(player_id) where kind='photo' and consumed_at is null;
--rollback drop table if exists camera_capture_items;
--rollback alter table camera_captures drop constraint if exists camera_captures_state_check, drop column if exists purchase_count, drop column if exists cleanup_attempted_at, drop column if exists deleted_at, drop column if exists abandoned_at, drop column if exists superseded_at, drop column if exists state;
