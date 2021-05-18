ALTER TABLE versions
    ADD COLUMN `api_aes_key` char(32) NOT NULL DEFAULT '' AFTER state;
