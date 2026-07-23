--liquibase formatted sql

--changeset pixels:furniture-seed-freeze-qa-usability-0040 context:development
update furniture_definitions set custom_params='120,180,300,600',updated_at=now()
where id=950017;

update furniture_items set extra_data='120',updated_at=now(),version=version+1
where id=961354 and room_id=151;

--rollback update furniture_definitions set custom_params='30,60,120,180,300,600',updated_at=now() where id=950017; update furniture_items set extra_data='30',updated_at=now(),version=version+1 where id=961354 and room_id=151;
