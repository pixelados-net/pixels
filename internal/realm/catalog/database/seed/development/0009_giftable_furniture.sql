--liquibase formatted sql

--changeset pixels:pixels-catalog-seed-development-0009-giftable-furniture context:development
update catalog_items
set giftable = true
where id in (1, 2, 3, 5, 6, 1002);
--rollback update catalog_items set giftable = false where id in (1, 2, 3, 5, 6);
