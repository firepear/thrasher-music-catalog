package querier

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

func TrkExists(db *sql.DB, trk string) bool {
	var r int
	db.QueryRow("select count(trk) from tracks where trk = ?", trk).Scan(&r)
	if r == 1 {
		return true
	}
	return false
}
