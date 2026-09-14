package config

import (
	"database/sql"
	"path/filepath"
	"testing"
)

func TestMigrateSeedsCategories(t *testing.T) {
	schemaPath, err := filepath.Abs("../../assets/database/schema.sql")
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("SCHEMA_PATH", schemaPath)
	db, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "forum.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := migrate(db); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"Общее", "Технологии", "Программирование", "Игры", "Музыка", "Спорт"} {
		var count int
		if err := db.QueryRow(`SELECT COUNT(*) FROM categories WHERE name = ?`, name).Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count != 1 {
			t.Fatalf("category %q: expected one row, got %d", name, count)
		}
	}
	if _, err := db.Exec(`INSERT INTO categories (name) VALUES ('Пользовательская')`); err != nil {
		t.Fatal(err)
	}
	var generalID, customID int
	if err := db.QueryRow(`SELECT id FROM categories WHERE name = 'Общее'`).Scan(&generalID); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT id FROM categories WHERE name = 'Пользовательская'`).Scan(&customID); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if err := migrate(db); err != nil {
			t.Fatal(err)
		}
	}
	var count, currentGeneralID, currentCustomID int
	if err := db.QueryRow(`SELECT COUNT(*) FROM categories`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT id FROM categories WHERE name = 'Общее'`).Scan(&currentGeneralID); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT id FROM categories WHERE name = 'Пользовательская'`).Scan(&currentCustomID); err != nil {
		t.Fatal(err)
	}
	if count != 7 || currentGeneralID != generalID || currentCustomID != customID {
		t.Fatalf("repeated migration changed categories: count=%d, IDs=%d/%d", count, currentGeneralID, currentCustomID)
	}
}

func TestMigratePreservesExistingCategories(t *testing.T) {
	schemaPath, err := filepath.Abs("../../assets/database/schema.sql")
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("SCHEMA_PATH", schemaPath)
	db, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "existing.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec(`CREATE TABLE categories (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL UNIQUE);
		INSERT INTO categories (id, name) VALUES (42, 'Общее'), (43, 'Старая категория');`); err != nil {
		t.Fatal(err)
	}
	if err := migrate(db); err != nil {
		t.Fatal(err)
	}
	var count, generalID, customID int
	if err := db.QueryRow(`SELECT COUNT(*) FROM categories`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT id FROM categories WHERE name = 'Общее'`).Scan(&generalID); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT id FROM categories WHERE name = 'Старая категория'`).Scan(&customID); err != nil {
		t.Fatal(err)
	}
	if count != 7 || generalID != 42 || customID != 43 {
		t.Fatalf("existing categories changed: count=%d, IDs=%d/%d", count, generalID, customID)
	}
}
