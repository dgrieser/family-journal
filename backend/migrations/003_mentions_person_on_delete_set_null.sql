SET @fk_person := (
  SELECT kcu.CONSTRAINT_NAME
  FROM information_schema.KEY_COLUMN_USAGE kcu
  WHERE kcu.TABLE_SCHEMA = DATABASE()
    AND kcu.TABLE_NAME = 'mentions'
    AND kcu.COLUMN_NAME = 'person_id'
    AND kcu.REFERENCED_TABLE_NAME = 'persons'
  LIMIT 1
);

SET @drop_fk_sql := IF(
  @fk_person IS NULL,
  'SELECT 1',
  CONCAT('ALTER TABLE mentions DROP FOREIGN KEY ', @fk_person)
);
PREPARE drop_fk_stmt FROM @drop_fk_sql;
EXECUTE drop_fk_stmt;
DEALLOCATE PREPARE drop_fk_stmt;

ALTER TABLE mentions
  DROP PRIMARY KEY,
  ADD COLUMN id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY FIRST,
  ADD UNIQUE KEY uniq_mentions_post_person (post_id, person_id);

ALTER TABLE mentions
  MODIFY person_id BIGINT NULL;

ALTER TABLE mentions
  ADD CONSTRAINT fk_mentions_person FOREIGN KEY (person_id) REFERENCES persons(id) ON DELETE SET NULL;
