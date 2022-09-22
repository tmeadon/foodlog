package data

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func migrate() error {
	files, err := getMigrationFiles()
	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	migrations, err := getActiveMigrations()
	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	for _, file := range files {
		filename := filepath.Base(file)

		if !contains(filename, migrations) {
			log.Printf("running migration %v", filename)

			query, err := os.ReadFile(file)
			if err != nil {
				return err
			}

			_ = DB.MustExec(string(query))
			writeActiveMigration(filename)
		}
	}

	return nil
}

func getMigrationFiles() ([]string, error) {
	files, err := filepath.Glob("./db/migrations/*.sql")
	if err != nil {
		return nil, fmt.Errorf("error getting migration files: %w", err)
	}

	sort.Strings(files)
	return files, nil
}

func getActiveMigrations() ([]string, error) {
	var migrations []string

	rows, err := DB.Queryx("select version from version order by id asc;")
	if err != nil {
		if strings.Contains(err.Error(), "no such table: version") {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read migrations from db: %w", err)
	}

	for rows.Next() {
		var m string

		err = rows.Scan(&m)
		if err != nil {
			return nil, fmt.Errorf("failed to read migrations from db: %w", err)
		}

		migrations = append(migrations, m)
	}

	return migrations, nil
}

func writeActiveMigration(migration string) {
	sql := "insert into version(version) values($1)"
	_ = DB.MustExec(sql, migration)
}

func contains(s string, a []string) bool {
	for _, k := range a {
		if s == k {
			return true
		}
	}
	return false
}
