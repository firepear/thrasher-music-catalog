package querier

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

type Handle struct {
	db       *sql.DB
	lastscan string
}

func New(dbfile string) *querier.Handle {
	MemDbURI = fmt.Sprintf("file:%s?mode=memory&cache=shared", DbName)

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		panic(err)
	}

	dbDisk := openDB(dbDiskPath)
	db := openDB(MemDbURI)
	err := backupDB(dbDisk, db)
}

func (h *Handle) TrkExists(trk string) bool {
	var r int
	db.QueryRow("select count(trk) from tracks where trk = ?", trk).Scan(&r)
	if r == 1 {
		return true
	}
	return false
}
