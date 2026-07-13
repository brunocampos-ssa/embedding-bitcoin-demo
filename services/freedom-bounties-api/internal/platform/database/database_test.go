package database

import (
	"context"
	"testing"
)

func TestSchemaAndSeedAreIdempotent(t *testing.T) {
	db, err := Open(context.Background(), ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err = Seed(context.Background(), db); err != nil {
		t.Fatal(err)
	}
	if err = Seed(context.Background(), db); err != nil {
		t.Fatal(err)
	}
	var n int
	if err = db.QueryRow(`SELECT COUNT(*) FROM bounties`).Scan(&n); err != nil || n != 3 {
		t.Fatalf("count=%d err=%v", n, err)
	}
}
