SET @has_posts_category := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS c
  WHERE c.TABLE_SCHEMA = DATABASE()
    AND c.TABLE_NAME = 'posts'
    AND c.COLUMN_NAME = 'category'
);

SET @has_posts_mood := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS c
  WHERE c.TABLE_SCHEMA = DATABASE()
    AND c.TABLE_NAME = 'posts'
    AND c.COLUMN_NAME = 'mood'
);

SET @drop_posts_legacy_columns_sql := CASE
  WHEN @has_posts_category = 1 AND @has_posts_mood = 1 THEN
    'ALTER TABLE posts DROP COLUMN category, DROP COLUMN mood'
  WHEN @has_posts_category = 1 THEN
    'ALTER TABLE posts DROP COLUMN category'
  WHEN @has_posts_mood = 1 THEN
    'ALTER TABLE posts DROP COLUMN mood'
  ELSE
    'SELECT 1'
END;

PREPARE drop_posts_legacy_columns_stmt FROM @drop_posts_legacy_columns_sql;
EXECUTE drop_posts_legacy_columns_stmt;
DEALLOCATE PREPARE drop_posts_legacy_columns_stmt;
