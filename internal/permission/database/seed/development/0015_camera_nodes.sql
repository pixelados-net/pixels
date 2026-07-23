--liquibase formatted sql

--changeset pixels:permission-seed-0015-camera-nodes labels:development
insert into permission_group_nodes(group_id,node,allowed)
select id,'camera.capture.use',true from permission_groups where name='member'
on conflict(group_id,node) do update set allowed=true;

insert into permission_group_nodes(group_id,node,allowed)
select id,node,true from permission_groups cross join(values
    ('camera.capture.use'),
    ('camera.settings.manage.any'),
    ('camera.gallery.moderate.any')
) nodes(node) where permission_groups.name='admin'
on conflict(group_id,node) do update set allowed=true;

insert into permission_group_nodes(group_id,node,allowed)
select id,'camera.gallery.moderate.any',true from permission_groups where name='moderator'
on conflict(group_id,node) do update set allowed=true;
--rollback delete from permission_group_nodes where node in('camera.capture.use','camera.settings.manage.any','camera.gallery.moderate.any');
