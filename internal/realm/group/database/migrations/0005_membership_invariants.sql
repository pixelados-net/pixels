--liquibase formatted sql

--changeset pixels:pixels-group-0005-membership-invariants splitStatements:false
alter table social_groups
    add constraint social_groups_active_owner_chk check(deactivated_at is not null or owner_player_id is not null);

create unique index social_group_members_single_owner_uidx
    on social_group_members(group_id) where role=0;

create function validate_social_group_owner() returns trigger language plpgsql as $$
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
              select 1 from social_group_members members
              where members.group_id=groups.id
                and members.player_id=groups.owner_player_id
                and members.role=0
          )
    ) then
        raise exception 'active social group % must have its owner membership',checked_group_id;
    end if;
    if exists(
        select 1
        from social_group_members members
        join social_groups groups on groups.id=members.group_id
        where members.group_id=checked_group_id
          and members.role=0
          and members.player_id<>groups.owner_player_id
    ) then
        raise exception 'social group % owner role does not match owner_player_id',checked_group_id;
    end if;
    return null;
end;
$$;

create constraint trigger social_groups_owner_consistency
after insert or update of owner_player_id,deactivated_at on social_groups
deferrable initially deferred for each row execute function validate_social_group_owner();

create constraint trigger social_group_members_owner_consistency
after insert or update or delete on social_group_members
deferrable initially deferred for each row execute function validate_social_group_owner();

--rollback drop trigger if exists social_group_members_owner_consistency on social_group_members; drop trigger if exists social_groups_owner_consistency on social_groups; drop function if exists validate_social_group_owner(); drop index if exists social_group_members_single_owner_uidx; alter table social_groups drop constraint if exists social_groups_active_owner_chk;
