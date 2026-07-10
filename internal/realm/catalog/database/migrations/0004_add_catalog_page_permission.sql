--liquibase formatted sql

--changeset pixels:pixels-catalog-0004-add-page-permission
alter table catalog_pages
    add column required_node text null,
    add constraint catalog_pages_required_node_length_chk
        check (required_node is null or char_length(required_node) between 1 and 160);
--rollback alter table catalog_pages drop constraint if exists catalog_pages_required_node_length_chk;
--rollback alter table catalog_pages drop column if exists required_node;
