package db

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jmoiron/sqlx"
)

func RunMigrations(db *sqlx.DB, migrationsDir string) error {
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version VARCHAR(255) PRIMARY KEY,
		applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`); err != nil {
		return fmt.Errorf("create schema_migrations table: %w", err)
	}

	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("read migrations directory %q: %w", migrationsDir, err)
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		files = append(files, entry.Name())
	}
	sort.Strings(files)

	for _, file := range files {
		var alreadyApplied int
		if err := db.Get(&alreadyApplied, "SELECT COUNT(*) FROM schema_migrations WHERE version = ?", file); err != nil {
			return fmt.Errorf("check migration %q: %w", file, err)
		}
		if alreadyApplied > 0 {
			continue
		}

		path := filepath.Join(migrationsDir, file)
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read migration %q: %w", path, err)
		}

		statements := splitSQLStatements(string(content))
		tx, err := db.Beginx()
		if err != nil {
			return fmt.Errorf("begin transaction for %q: %w", file, err)
		}

		for idx, stmt := range statements {
			if _, err := tx.Exec(stmt); err != nil {
				_ = tx.Rollback()
				return fmt.Errorf("execute migration %q statement %d: %w", file, idx+1, err)
			}
		}

		if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES (?)", file); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("record migration %q: %w", file, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %q: %w", file, err)
		}
	}

	return nil
}

func splitSQLStatements(sql string) []string {
	var (
		stmts            []string
		builder          strings.Builder
		inSingleQuote    bool
		inDoubleQuote    bool
		inBacktickQuote  bool
		inLineComment    bool
		inBlockComment   bool
		prevWasBackslash bool
		runes            = []rune(sql)
	)

	for i := 0; i < len(runes); i++ {
		ch := runes[i]
		next := rune(0)
		if i+1 < len(runes) {
			next = runes[i+1]
		}

		if inLineComment {
			if ch == '\n' {
				inLineComment = false
			}
			continue
		}
		if inBlockComment {
			if ch == '*' && next == '/' {
				inBlockComment = false
				i++
			}
			continue
		}
		if !inSingleQuote && !inDoubleQuote && !inBacktickQuote {
			if ch == '-' && next == '-' {
				inLineComment = true
				i++
				continue
			}
			if ch == '#' {
				inLineComment = true
				continue
			}
			if ch == '/' && next == '*' {
				inBlockComment = true
				i++
				continue
			}
		}

		switch ch {
		case '\'':
			if !inDoubleQuote && !inBacktickQuote && !prevWasBackslash {
				inSingleQuote = !inSingleQuote
			}
		case '"':
			if !inSingleQuote && !inBacktickQuote && !prevWasBackslash {
				inDoubleQuote = !inDoubleQuote
			}
		case '`':
			if !inSingleQuote && !inDoubleQuote {
				inBacktickQuote = !inBacktickQuote
			}
		case ';':
			if !inSingleQuote && !inDoubleQuote && !inBacktickQuote {
				stmt := strings.TrimSpace(builder.String())
				if stmt != "" {
					stmts = append(stmts, stmt)
				}
				builder.Reset()
				prevWasBackslash = false
				continue
			}
		}

		builder.WriteRune(ch)
		prevWasBackslash = ch == '\\' && !prevWasBackslash
		if ch != '\\' {
			prevWasBackslash = false
		}
	}

	last := strings.TrimSpace(builder.String())
	if last != "" {
		stmts = append(stmts, last)
	}
	return stmts
}
