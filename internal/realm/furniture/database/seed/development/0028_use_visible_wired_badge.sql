--liquibase formatted sql

--changeset pixels:pixels-furniture-seed-development-0028-use-visible-wired-badge context:development
update room_wired_settings
set string_param='ADM',updated_at=now()
where item_id in (420011,420021,420131);
