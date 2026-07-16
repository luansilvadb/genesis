// Package database provides database bootstrap and migration utilities.
//
// Migration strategy: GORM AutoMigrate (in cmd/server/main.go) is the primary
// schema management tool. Raw SQL migrations in the migrations/ directory are
// supplementary for operations AutoMigrate cannot express (e.g. adding UNIQUE
// constraints, creating indexes). RunSQLMigrations is available for manual or
// CI-driven execution of those raw SQL files.
//
// splitSQL handles PostgreSQL dollar-quoting ($$) so DO blocks and function
// bodies in migration SQL files are kept intact as single statements.
package database

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gorm.io/gorm"
)

func RunSQLMigrations(db *gorm.DB, migrationsDir string) error {
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory %s: %w", migrationsDir, err)
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)

	for _, fname := range files {
		path := filepath.Join(migrationsDir, fname)
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", path, err)
		}

		statements := splitSQL(string(content))
		for i, stmt := range statements {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				continue
			}
			if err := db.Exec(stmt).Error; err != nil {
				return fmt.Errorf("migration %s statement %d failed: %w", fname, i+1, err)
			}
		}
	}

	return nil
}

// splitSQL splits a raw SQL string into individual statements separated by
// semicolons, respecting PostgreSQL dollar-quoting ($$ … $$) and single-quoted
// strings so that DO blocks and function bodies remain intact.
//
// The character-by-character state machine is the idiomatic pattern for this
// problem. Extracting the dollar-quote state tracking into a separate type
// would fragment the logic across multiple functions without simplifying the
// underlying decision count — the three states (normal, in-string, in-dollar)
// are inherently coupled to each character transition. The cyclomatic
// complexity (~13) is justified domain complexity, not accidental.
func splitSQL(sql string) []string {
	var result []string
	var current strings.Builder
	inString := false
	inDollar := false

	for i, ch := range sql {
		switch {
		case ch == '$' && !inString && i+1 < len(sql) && sql[i+1] == '$':
			inDollar = !inDollar
			current.WriteRune(ch)
		case ch == '\'':
			if !inDollar {
				inString = !inString
			}
			current.WriteRune(ch)
		case ch == ';':
			if !inString && !inDollar {
				result = append(result, current.String())
				current.Reset()
			} else {
				current.WriteRune(ch)
			}
		default:
			current.WriteRune(ch)
		}
	}

	remaining := strings.TrimSpace(current.String())
	if remaining != "" {
		result = append(result, remaining)
	}

	return result
}
