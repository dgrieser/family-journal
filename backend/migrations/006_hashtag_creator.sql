ALTER TABLE hashtags ADD COLUMN created_by_user_id BIGINT NULL;

UPDATE hashtags h
SET created_by_user_id = (
    SELECT p.user_id
    FROM post_hashtags ph
    JOIN posts p ON p.id = ph.post_id
    WHERE ph.hashtag_id = h.id
    ORDER BY p.created_at ASC
    LIMIT 1
)
WHERE created_by_user_id IS NULL;
