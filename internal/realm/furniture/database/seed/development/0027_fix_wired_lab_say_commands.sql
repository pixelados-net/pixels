--liquibase formatted sql

--changeset pixels:pixels-furniture-seed-development-0027-fix-wired-lab-say-commands context:development
-- Keep substring-matched WIRED QA commands mutually exclusive within each room.
insert into room_wired_settings(item_id,int_params,string_param,selection_mode,delay_pulses)
values
    (420001,'[]','Actor lab ready. Commands: badgeyes, badgeno, groupyes, groupno, effecton, haseffect, noeffect, handon, hashand, countone, notcount, date, orcheck.',0,0),
    (420010,'[]','badgeyes',0,0),
    (420020,'[]','badgeno',0,0),
    (420030,'[]','groupyes',0,0),
    (420040,'[]','groupno',0,0),
    (421040,'[]','stackyes',0,0),
    (421050,'[]','stackno',0,0),
    (421110,'[]','elapsed',0,0),
    (423070,'[]','playerscore',0,0),
    (423120,'[]','teamyes',0,0),
    (423130,'[]','teamno',0,0)
on conflict(item_id) do update set
    int_params=excluded.int_params,
    string_param=excluded.string_param,
    selection_mode=excluded.selection_mode,
    delay_pulses=excluded.delay_pulses,
    updated_at=now();
