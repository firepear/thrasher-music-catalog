package tmc

import (
	"database/sql"
	"fmt"
)

var (
	migrations []migration
)

type migration struct {
	migr func() error
	desc string
}

func init() {
	migrations = []migration{
		migration{migr: migr00,	desc: "Add migration support; Update ctimes"},
	}
}

// migratedb is the driver function for database migrations. it is
// called from New().
func migratedb(db *sql.DB) (error) {
	var dbver int64

	db.QueryRow("SELECT version FROM meta").Scan(&dbver)
	fmt.Println(dbver)

	return nil
}


//////////////////////////////////////////////////////// migration funcs

func migr00() error {
	return nil
}
