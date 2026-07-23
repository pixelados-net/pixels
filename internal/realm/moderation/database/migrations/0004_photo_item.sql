--liquibase formatted sql

--changeset pixels:pixels-moderation-0004-photo-item
alter table moderation_issues add column photo_item_id bigint null references furniture_items(id) on delete set null;
create index moderation_issues_photo_item_idx on moderation_issues(photo_item_id) where photo_item_id is not null;
--rollback drop index if exists moderation_issues_photo_item_idx; alter table moderation_issues drop column if exists photo_item_id;
