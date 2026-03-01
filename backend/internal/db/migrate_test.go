package db

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestFreshSchemaStaysCompatibleWithFollowUpMigrations(t *testing.T) {
	initSQL := mustReadMigration(t, "001_init.sql")
	if !strings.Contains(initSQL, "CREATE TABLE IF NOT EXISTS mentions (\n  id BIGINT AUTO_INCREMENT PRIMARY KEY,") {
		t.Fatal("expected 001_init.sql to create mentions.id in the base schema")
	}
	if strings.Contains(initSQL, "category ") || strings.Contains(initSQL, "mood ") {
		t.Fatal("expected 001_init.sql to omit legacy posts.category and posts.mood columns")
	}
	if strings.Contains(initSQL, "url VARCHAR") {
		t.Fatal("expected 001_init.sql to omit legacy attachments.url column")
	}

	migration003 := mustReadMigration(t, "003_mentions_person_on_delete_set_null.sql")
	for _, snippet := range []string{
		"FROM information_schema.COLUMNS c",
		"c.TABLE_NAME = 'mentions'",
		"c.COLUMN_NAME = 'id'",
		"FROM information_schema.STATISTICS s",
		"s.INDEX_NAME = 'uniq_mentions_post_person'",
		"SET @add_mentions_id_sql := IF(",
		"SET @add_mentions_post_person_uniq_sql := IF(",
	} {
		if !strings.Contains(migration003, snippet) {
			t.Fatalf("expected 003_mentions_person_on_delete_set_null.sql to contain %q", snippet)
		}
	}

	migration004 := mustReadMigration(t, "004_remove_post_category_mood.sql")
	for _, snippet := range []string{
		"c.TABLE_NAME = 'posts'",
		"c.COLUMN_NAME = 'category'",
		"c.COLUMN_NAME = 'mood'",
		"SET @drop_posts_legacy_columns_sql := CASE",
	} {
		if !strings.Contains(migration004, snippet) {
			t.Fatalf("expected 004_remove_post_category_mood.sql to contain %q", snippet)
		}
	}

	migration005 := mustReadMigration(t, "005_remove_attachment_url_column.sql")
	for _, snippet := range []string{
		"c.TABLE_NAME = 'attachments'",
		"c.COLUMN_NAME = 'url'",
		"SET @drop_attachments_url_sql := IF(",
	} {
		if !strings.Contains(migration005, snippet) {
			t.Fatalf("expected 005_remove_attachment_url_column.sql to contain %q", snippet)
		}
	}
}

func TestSplitSQLStatementsHandlesGuardedMigrations(t *testing.T) {
	tests := []struct {
		file          string
		wantStatements int
	}{
		{file: "003_mentions_person_on_delete_set_null.sql", wantStatements: 17},
		{file: "004_remove_post_category_mood.sql", wantStatements: 6},
		{file: "005_remove_attachment_url_column.sql", wantStatements: 5},
	}

	for _, tc := range tests {
		t.Run(tc.file, func(t *testing.T) {
			statements := splitSQLStatements(mustReadMigration(t, tc.file))
			if len(statements) != tc.wantStatements {
				t.Fatalf("expected %d statements, got %d", tc.wantStatements, len(statements))
			}
			for i, stmt := range statements {
				if strings.TrimSpace(stmt) == "" {
					t.Fatalf("statement %d is empty", i+1)
				}
			}
		})
	}
}

func mustReadMigration(t *testing.T, file string) string {
	t.Helper()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve current file path")
	}

	path := filepath.Join(filepath.Dir(currentFile), "..", "..", "migrations", file)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", file, err)
	}
	return string(data)
}
