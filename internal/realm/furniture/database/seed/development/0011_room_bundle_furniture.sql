--liquibase formatted sql

--changeset pixels:pixels-furniture-seed-development-0011-room-bundle-furniture context:development
delete from furniture_items where id between 21001 and 21024;
insert into furniture_items (id,definition_id,owner_player_id,room_id,x,y,z,rotation,extra_data,limited_edition_number,metadata)
overriding system value values
    (21001,3,1,100,6,2,0,0,'0',77,'{"bundle":"starter_loft"}'),
    (21002,1,1,100,7,6,0,0,'0',null,'{"bundle":"starter_loft"}'),
    (21003,2,1,100,6,6,0,2,'0',null,'{"bundle":"starter_loft"}'),
    (21004,2,1,100,9,6,0,6,'0',null,'{"bundle":"starter_loft"}'),
    (21005,25,1,100,7,3,0,0,'1',null,'{"bundle":"starter_loft"}'),
    (21006,26,1,100,10,3,0,0,'1',null,'{"bundle":"starter_loft"}'),
    (21007,3,1,100,8,10,0,0,'0',null,'{"bundle":"starter_loft"}'),
    (21008,10,1,101,3,6,0,0,'5',null,'{"bundle":"interactive_lounge"}'),
    (21009,10,1,101,4,6,0,0,'2',null,'{"bundle":"interactive_lounge"}'),
    (21010,11,1,101,6,2,0,0,'0',null,'{"bundle":"interactive_lounge"}'),
    (21011,27,1,101,8,2,0,0,'4',null,'{"bundle":"interactive_lounge"}'),
    (21012,29,1,101,2,8,0,0,'1',null,'{"bundle":"interactive_lounge"}'),
    (21013,28,1,101,3,8,0,0,'0',null,'{"bundle":"interactive_lounge"}'),
    (21014,6,1,101,5,9,0,0,'1',null,'{"bundle":"interactive_lounge"}'),
    (21015,7,1,101,7,9,0,0,'0',null,'{"bundle":"interactive_lounge"}'),
    (21016,3,1,101,9,7,0,0,'0',null,'{"bundle":"interactive_lounge"}'),
    (21017,5,1,102,6,5,0,0,'0',null,'{"bundle":"cozy_bedroom"}'),
    (21018,4,1,102,9,5,0,0,'0',null,'{"bundle":"cozy_bedroom"}'),
    (21019,26,1,102,6,9,0,0,'1',null,'{"bundle":"cozy_bedroom"}'),
    (21020,3,1,102,8,9,0,0,'0',null,'{"bundle":"cozy_bedroom"}'),
    (21021,34,1,102,10,7,0,0,'1',null,'{"bundle":"cozy_bedroom"}'),
    (21022,35,1,102,10,9,0,0,'0',null,'{"bundle":"cozy_bedroom"}');
select setval(pg_get_serial_sequence('furniture_items','id'),greatest((select max(id) from furniture_items),1));
--rollback delete from furniture_items where id between 21001 and 21024;
