--liquibase formatted sql

--changeset pixels:pixels-catalog-0008-create-vouchers
create table catalog_vouchers (
    id bigint generated always as identity primary key,
    code text not null,
    cost_credits bigint not null default 0,
    cost_points bigint not null default 0,
    points_type integer not null default -1,
    catalog_item_id bigint null references catalog_items(id),
    redemption_cap integer null,
    per_player_cap integer not null default 1,
    enabled boolean not null default true,
    created_at timestamptz not null default now(),
    expires_at timestamptz null,
    constraint catalog_vouchers_code_length_chk check (char_length(code) between 4 and 32),
    constraint catalog_vouchers_caps_chk check ((redemption_cap is null or redemption_cap > 0) and per_player_cap > 0)
);
create unique index catalog_vouchers_code_uidx on catalog_vouchers (upper(code));
create table catalog_voucher_redemptions (
    voucher_id bigint not null references catalog_vouchers(id),
    player_id bigint not null references players(id),
    redeemed_at timestamptz not null default now(),
    primary key (voucher_id, player_id)
);
--rollback drop table if exists catalog_voucher_redemptions; drop table if exists catalog_vouchers;
