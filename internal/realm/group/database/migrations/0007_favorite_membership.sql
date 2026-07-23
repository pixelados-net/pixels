--liquibase formatted sql

--changeset pixels:pixels-group-0007-favorite-membership
alter table player_social_group_preferences
    add constraint player_social_group_favorite_membership_fk
    foreign key(favorite_group_id,player_id)
    references social_group_members(group_id,player_id)
    deferrable initially deferred;

--rollback alter table player_social_group_preferences drop constraint if exists player_social_group_favorite_membership_fk;
