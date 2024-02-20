package database_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/catalogfi/indexer/database"
	"go.uber.org/zap"
)

func TestRocksDB(t *testing.T) {
	path := t.TempDir()
	db, err := database.NewRocksDB(path, zap.NewNop())
	if err != nil {
		t.Fatal("expected database error to be nil")
	}
	if db == nil {
		t.Fatal("db is nil")
	}

	// generate a random sha256 hash
	value := []byte("value fjkhf djwhbds jdbhjd djdbw ddbhd jdhwdbhd ehjdeuhdyed edbhuebdhedhe deuhdhedfghed ehd ghedghe dehudghedhe dheudeg dge gdegd ge dgefg ehuebfhed eg dge gdehdbedugebihdegdg dg ehbihefuydvegufhe dehdijeduevfgvegd")
	timeNow := time.Now()
	for i := 0; i < 1; i++ {
		err = db.Put(fmt.Sprintf("key %v", i), value)
		if err != nil {
			t.Fatal("expected database error to be nil")
		}
	}
	t.Log("Put 10000 times:", time.Since(timeNow))

	timeNow = time.Now()
	for i := 0; i < 10000; i++ {
		_, err = db.Get(fmt.Sprintf("key %v", i))
		if err != nil {
			t.Fatal("expected database error to be nil")
		}
	}
	t.Log("Get 10000 times:", time.Since(timeNow))
}
