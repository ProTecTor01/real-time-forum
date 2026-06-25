package config

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"src/internal/api"
	"src/internal/sessions"
	"src/internal/utils"
	"src/internal/ws"

	_ "github.com/mattn/go-sqlite3"
)

//--------------------------------------------------------------------------------------

const (
	defaultPort      = 8080
	defaultDBPath    = "./assets/database/forum.db"
	DefaultStaticDir = "./assets/static/"
)

//--------------------------------------------------------------------------------------

type Config struct {
	DB   *sql.DB
	Port int
}

//--------------------------------------------------------------------------------------

func Setup() (*Config, error) {
	dbPath := utils.Getenv("DB_PATH", defaultDBPath)
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open db: %v", err)
	}

	if err := migrate(db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("migrate db: %v", err)
	}

	sessionManager := sessions.NewSessionManager(db, sessions.DefaultSessionTTL)
	hub := ws.NewHub()
	go hub.Run()

	apiHandler := api.NewAPI(db, sessionManager, hub)

	staticDir := utils.Getenv("STATIC_DIR", DefaultStaticDir)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))
	http.Handle("/assets/uploads/", http.StripPrefix("/assets/uploads/", http.FileServer(http.Dir("./assets/uploads"))))
	http.HandleFunc("/", apiHandler.Index)
	http.HandleFunc("/ws", apiHandler.WS)

	http.HandleFunc("/api/register", apiHandler.Register)
	http.HandleFunc("/api/login", apiHandler.Login)
	http.HandleFunc("/api/logout", apiHandler.Logout)
	http.HandleFunc("/api/me", apiHandler.Me)

	http.HandleFunc("/api/posts", apiHandler.Posts)
	http.HandleFunc("/api/posts/", apiHandler.PostByID)
	http.HandleFunc("/api/comments", apiHandler.Comments)
	http.HandleFunc("/api/categories", apiHandler.Categories)
	http.HandleFunc("/api/likes", apiHandler.Likes)

	http.HandleFunc("/api/users", apiHandler.Users)
	http.HandleFunc("/api/chats", apiHandler.ChatList)
	http.HandleFunc("/api/messages", apiHandler.Messages)
	http.HandleFunc("/api/profiles/", apiHandler.Profile)
	http.HandleFunc("/api/me/profile", apiHandler.MyProfile)
	http.HandleFunc("/api/upload", apiHandler.UploadImage)

	cfg := &Config{
		DB:   db,
		Port: utils.GetIntEnv("PORT", defaultPort),
	}
	return cfg, nil
}

//--------------------------------------------------------------------------------------

func (c *Config) Close() {
	if err := c.DB.Close(); err != nil {
		log.Printf("Error closing database: %v", err)
	}
}

//--------------------------------------------------------------------------------------

func migrate(db *sql.DB) error {
	schemaPath := os.Getenv("SCHEMA_PATH")
	if schemaPath == "" {
		schemaPath = "./assets/database/schema.sql"
	}
	schema, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("read schema %s: %w", schemaPath, err)
	}

	if _, err = db.Exec(string(schema)); err != nil {
		return fmt.Errorf("apply schema %s: %w", schemaPath, err)
	}

	migrations := []struct {
		table      string
		column     string
		definition string
	}{
		{"posts", "author", "author TEXT"},
		{"users", "first_name", "first_name TEXT NOT NULL DEFAULT ''"},
		{"users", "last_name", "last_name TEXT NOT NULL DEFAULT ''"},
		{"users", "age", "age INTEGER NOT NULL DEFAULT 18"},
		{"users", "gender", "gender TEXT NOT NULL DEFAULT 'other'"},
		{"messages", "image_path", "image_path TEXT"},
		{"users", "avatar_path", "avatar_path TEXT"},
		{"posts", "image_path", "image_path TEXT"},
	}
	for _, migration := range migrations {
		if err := ensureColumn(db, migration.table, migration.column, migration.definition); err != nil {
			return err
		}
	}

	return nil
}

func ensureColumn(db *sql.DB, table, column, definition string) error {
	exists, err := columnExists(db, table, column)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	if _, err := db.Exec(fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s", table, definition)); err != nil {
		return fmt.Errorf("add column %s.%s: %w", table, column, err)
	}
	return nil
}

func columnExists(db *sql.DB, table, column string) (bool, error) {
	rows, err := db.Query(fmt.Sprintf("PRAGMA table_info(%s)", table))
	if err != nil {
		return false, fmt.Errorf("inspect table %s: %w", table, err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			cid          int
			name         string
			columnType   string
			notNull      int
			defaultValue sql.NullString
			primaryKey   int
		)
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &primaryKey); err != nil {
			return false, fmt.Errorf("scan column info for %s: %w", table, err)
		}
		if name == column {
			return true, nil
		}
	}
	if err := rows.Err(); err != nil {
		return false, fmt.Errorf("iterate columns for %s: %w", table, err)
	}
	return false, nil
}
