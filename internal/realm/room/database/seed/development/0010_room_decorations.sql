--liquibase formatted sql

--changeset pixels:pixels-room-seed-development-0010-room-decorations context:development
update rooms set floor_paint='101',wallpaper='101',landscape='1.1' where id=100;
update rooms set floor_paint='110',wallpaper='110',landscape='2.1' where id=101;
update rooms set floor_paint='201',wallpaper='201',landscape='3.1' where id=102;
insert into room_dimmer_presets(room_id,preset_id,background_only,color,brightness,selected,enabled)
values (100,1,false,'#74F5F5',180,true,true),(100,2,true,'#E759DE',220,false,false),(100,3,false,'#000000',255,false,false)
on conflict(room_id,preset_id) do update set background_only=excluded.background_only,color=excluded.color,brightness=excluded.brightness,selected=excluded.selected,enabled=excluded.enabled;
--rollback delete from room_dimmer_presets where room_id=100; update rooms set floor_paint='0.0',wallpaper='0.0',landscape='0.0' where id in (100,101,102);
