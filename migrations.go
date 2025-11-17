package tmc

import (
	"database/sql"

	//sqlite "github.com/mattn/go-sqlite3"
)

var (
	migrations []Migration
)

type Migration struct {
	migr func() error
	desc string
}

func init() {
	migrations = []Migration{
		Migration{migr: migr00,	desc: "Add migration support; Update ctimes"},
	}
}


func Migrate(db *sql.DB) (error) {
	return nil
}


//////////////////////////////////////////////////////// migration funcs

func migr00() error {
	return nil
}
