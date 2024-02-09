package database_test

import (
	"fmt"
	"testing"

	"github.com/catalogfi/indexer/database"
)

func TestMDBX(t *testing.T) {
	path := t.TempDir()
	dbName := t.Name()

	db, err := database.NewMDBX(path, dbName)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	t.Run("should be able to put and get", func(t *testing.T) {
		key := t.Name()
		value := "values"
		if err := db.Put(key, []byte(value)); err != nil {
			t.Fatal(err)
		}
		got, err := db.Get(key)
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != value {
			t.Fatalf("expected %s, got %s", value, got)
		}
	})

	t.Run("should be able to delete", func(t *testing.T) {
		key := t.Name()
		value := "Delvalues"
		if err := db.Put(key, []byte(value)); err != nil {
			t.Fatal(err)
		}
		if err := db.Delete(key); err != nil {
			t.Fatal(err)
		}
		got, err := db.Get(key)
		fmt.Println(string(got))
		if err == nil {
			t.Fatalf("expected error, got %s", got)
		}
	})
}
