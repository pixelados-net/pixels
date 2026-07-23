--liquibase formatted sql

--changeset pixels:pixels-furniture-0010-recover-pending-postits
update furniture_items item
set extra_data = 'FFFF33 ',
    metadata = jsonb_set(item.metadata, '{migration_0010_previous_postit_data}', to_jsonb(item.extra_data), true),
    updated_at = now(),
    version = item.version + 1
from furniture_definitions definition
where item.definition_id = definition.id
  and definition.interaction_type = 'postit'
  and item.room_id is not null
  and item.deleted_at is null
  and length(item.extra_data) < 6;

--rollback update furniture_items set extra_data = metadata->>'migration_0010_previous_postit_data', metadata = metadata-'migration_0010_previous_postit_data', updated_at = now(), version = version + 1 where jsonb_exists(metadata, 'migration_0010_previous_postit_data');
