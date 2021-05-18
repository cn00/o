ALTER TABLE file_urls
    ADD COLUMN `state` int default 2 AFTER url;

ALTER TABLE resource_urls
    ADD COLUMN `state` int default 2 AFTER url;

ALTER TABLE versions
    ADD COLUMN `state` int default 2 AFTER env_id;

ALTER TABLE versions
    ADD COLUMN `upd_datetime` datetime default NOW() AFTER state;


-- DROP COLUMN
ALTER TABLE file_urls DROP COLUMN `state`;
ALTER TABLE resource_urls DROP COLUMN `state`;
ALTER TABLE versions DROP COLUMN `state`;
ALTER TABLE versions DROP COLUMN `upd_datetime`;
