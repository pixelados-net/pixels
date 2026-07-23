--liquibase formatted sql

--changeset pixels:pixels-player-seed-development-0005-user-remaining context:development
--validCheckSum:ANY
insert into player_settings(player_id,volume_system,volume_furniture,volume_trax,old_chat,camera_follow_blocked,safety_locked)
values (1,80,70,60,false,false,false),(2,100,90,80,true,false,false),(3,100,100,100,false,false,true),(4,50,50,50,false,true,false)
on conflict(player_id) do update set volume_system=excluded.volume_system,volume_furniture=excluded.volume_furniture,volume_trax=excluded.volume_trax,old_chat=excluded.old_chat,camera_follow_blocked=excluded.camera_follow_blocked,safety_locked=excluded.safety_locked;

insert into player_profile_tags(player_id,position,tag)
values (1,1,'pixels'),(1,2,'builder'),(1,3,'host'),(2,1,'rooms')
on conflict(player_id,position) do update set tag=excluded.tag;

insert into player_wardrobe_outfits(player_id,slot_id,figure,gender)
values
(1,1,'hr-100.hd-180-1.ch-210-66.lg-270-82.sh-290-80','M'),
(1,2,'hr-100.hd-180-1.ch-255-81.lg-280-64.sh-305-62','M'),
(2,1,'hr-515-45.hd-600-1.ch-665-92.lg-700-64.sh-735-68','F')
on conflict(player_id,slot_id) do update set figure=excluded.figure,gender=excluded.gender,updated_at=now();

update player_profiles set allow_name_change=(player_id=1) where player_id in (1,2,3,4);

insert into player_respect_grants(actor_player_id,target_player_id,grant_date,source)
values (4,1,current_date,'seed'),(4,2,current_date,'seed'),(4,3,current_date,'seed')
on conflict do nothing;

--rollback delete from player_respect_grants where actor_player_id=4 and grant_date=current_date and source='seed'; delete from player_wardrobe_outfits where player_id in (1,2); delete from player_profile_tags where player_id in (1,2); delete from player_settings where player_id in (1,2,3,4); update player_profiles set allow_name_change=false where player_id in (1,2,3,4);
