package updater

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

// MigrateDB is the driver function for database migrations.
func MigrateDB(db *sql.DB) (error) {
	var dbver int64

	db.QueryRow("SELECT version FROM meta").Scan(&dbver)
	fmt.Println("dbver ", dbver)

	return nil
}


//////////////////////////////////////////////////////// migration funcs

func migr00() error {
	return nil
}
