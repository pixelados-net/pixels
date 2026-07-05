--liquibase formatted sql

--changeset pixels:pixels-shared-0001-extensions
create extension if not exists pgcrypto;
--rollback drop extension if exists pgcrypto;
