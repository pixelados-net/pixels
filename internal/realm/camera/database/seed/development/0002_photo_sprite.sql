--liquibase formatted sql

--changeset pixels:pixels-camera-seed-development-0002-photo-sprite context:development
update furniture_definitions
set sprite_id=4597,
    name='external_image_wallitem_poster_small',
    public_name='Foto',
    interaction_type='external_image',
    interaction_modes_count=2,
    allow_recycle=false
where id=940001;
--rollback update furniture_definitions set sprite_id=4536 where id=940001;
