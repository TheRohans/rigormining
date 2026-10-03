package repository

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"

	"gitlab.com/robrohan/rigormining/internals/models"
)

// testRepo opens a fresh, fully migrated database: SQLite by default, or
// Postgres when RM_TEST_POSTGRES holds a connection string (see the
// Makefile's test-postgres target). The Postgres database's public schema
// is wiped first - point it at a throwaway database only.
//
// SQLite runs with foreign keys enforced, which it doesn't do by default,
// so it fails the same way Postgres would on a dangling reference.
func testRepo(t *testing.T) *DataRepository {
	t.Helper()

	driver, conn := "sqlite3", filepath.Join(t.TempDir(), "test.db")+"?_foreign_keys=on"
	if dsn := os.Getenv("RM_TEST_POSTGRES"); dsn != "" {
		driver, conn = "postgres", dsn
		reset, err := sqlx.Open("postgres", dsn)
		if err != nil {
			t.Fatal(err)
		}
		_, err = reset.Exec(`DROP SCHEMA public CASCADE; CREATE SCHEMA public`)
		reset.Close()
		if err != nil {
			t.Fatal(err)
		}
	}

	// Migrations are read from ./migrations, relative to the backend root.
	wd, _ := os.Getwd()
	if err := os.Chdir("../.."); err != nil {
		t.Fatal(err)
	}
	db, err := OpenDatabase(driver, conn, "test")
	os.Chdir(wd)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return Attach("test", db, driver)
}

func testUser(t *testing.T, r *DataRepository, email string) string {
	t.Helper()
	if err := r.UpsertUser(models.NewUser(email, email, "")); err != nil {
		t.Fatal(err)
	}
	u, err := r.GetUser(email)
	if err != nil {
		t.Fatal(err)
	}
	return u.UUID
}

func testItem(t *testing.T, r *DataRepository, userId string, tags ...string) string {
	t.Helper()
	item := &models.LibraryItem{
		UUID: uuid.New().String(), UserId: userId, Title: "paper",
		AddedDate: time.Now().UTC().Format(time.RFC3339),
	}
	if err := r.CreateItem(item); err != nil {
		t.Fatal(err)
	}
	for _, tag := range tags {
		if err := r.AttachTag(item.UUID, tag); err != nil {
			t.Fatal(err)
		}
	}
	return item.UUID
}

func tagNames(t *testing.T, r *DataRepository, itemId string) []string {
	t.Helper()
	tags, err := r.GetTagsForItem(itemId)
	if err != nil {
		t.Fatal(err)
	}
	names := []string{}
	for _, tag := range tags {
		names = append(names, tag.Name)
	}
	sort.Strings(names)
	return names
}
