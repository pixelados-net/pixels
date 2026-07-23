--liquibase formatted sql

--changeset pixels:pixels-permission-seed-development-0012-pet-nodes context:development
insert into permission_group_nodes(group_id,node,allowed)
select id,node,true
from permission_groups
cross join (values
    ('pet.manage.any'),
    ('pet.place.any'),
    ('pet.room.limit.bypass'),
    ('pet.inventory.limit.bypass'),
    ('pet.respect.limit.bypass'),
    ('pet.lifecycle.manage'),
    ('pet.move.any')
) as pet_nodes(node)
where permission_groups.name='admin'
on conflict(group_id,node) do update set allowed=excluded.allowed;

insert into permission_group_nodes(group_id,node,allowed)
select id,node,true
from permission_groups
cross join (values
    ('pet.place.any'),
    ('pet.move.any')
) as pet_nodes(node)
where permission_groups.name='moderator'
on conflict(group_id,node) do update set allowed=excluded.allowed;

--rollback delete from permission_group_nodes where node in ('pet.manage.any','pet.place.any','pet.room.limit.bypass','pet.inventory.limit.bypass','pet.respect.limit.bypass','pet.lifecycle.manage','pet.move.any');
