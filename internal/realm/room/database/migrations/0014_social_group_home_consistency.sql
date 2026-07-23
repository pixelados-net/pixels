--liquibase formatted sql

--changeset pixels:pixels-room-0014-social-group-home-consistency splitStatements:false
create function validate_social_group_home_room() returns trigger language plpgsql as $$
declare
    checked_group_id bigint;
begin
    if tg_table_name='social_groups' then
        checked_group_id := coalesce(new.id,old.id);
    else
        checked_group_id := coalesce(new.group_id,old.group_id);
    end if;
    if exists(
        select 1
        from social_groups groups
        where groups.id=checked_group_id
          and groups.deactivated_at is null
          and not exists(
              select 1 from room_social_groups bindings
              where bindings.group_id=groups.id
                and bindings.room_id=groups.home_room_id
          )
    ) then
        raise exception 'active social group % must match its room binding',checked_group_id;
    end if;
    return null;
end;
$$;

create constraint trigger social_groups_home_consistency
after insert or update of home_room_id,deactivated_at on social_groups
deferrable initially deferred for each row execute function validate_social_group_home_room();

create constraint trigger room_social_groups_home_consistency
after insert or update or delete on room_social_groups
deferrable initially deferred for each row execute function validate_social_group_home_room();

--rollback drop trigger if exists room_social_groups_home_consistency on room_social_groups; drop trigger if exists social_groups_home_consistency on social_groups; drop function if exists validate_social_group_home_room();
