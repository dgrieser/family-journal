SET @has_attachments_url := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS c
  WHERE c.TABLE_SCHEMA = DATABASE()
    AND c.TABLE_NAME = 'attachments'
    AND c.COLUMN_NAME = 'url'
);

SET @drop_attachments_url_sql := IF(
  @has_attachments_url = 1,
  'ALTER TABLE attachments DROP COLUMN url',
  'SELECT 1'
);

PREPARE drop_attachments_url_stmt FROM @drop_attachments_url_sql;
EXECUTE drop_attachments_url_stmt;
DEALLOCATE PREPARE drop_attachments_url_stmt;
