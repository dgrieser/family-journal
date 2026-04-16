ALTER TABLE hashtags ADD COLUMN name_lower VARCHAR(120) NOT NULL DEFAULT '';
UPDATE hashtags SET name_lower = LOWER(name);
ALTER TABLE hashtags ADD UNIQUE KEY uniq_hashtag_name_lower (name_lower);
