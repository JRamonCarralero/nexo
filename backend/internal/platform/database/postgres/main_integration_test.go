//go:build integration

package postgres_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/JRamonCarralero/nexo/backend/internal/platform/database/postgres/pgtest"
)

var testDB *pgtest.Database

func TestMain(m *testing.M) {
	ctx := context.Background()
	var err error
	testDB, err = pgtest.Start(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	code := m.Run()
	testDB.Stop(ctx)
	os.Exit(code)
}
