package "tmc"

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

type Catalog struct {
	db       *sql.DB
	lastscan int
}

func New() *tmc.Catalog {
	MemDbURI = fmt.Sprintf("file:%s?mode=memory&cache=shared", DbName)

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		panic(err)
	}

	dbDisk := openDB(dbDiskPath)
	db := openDB(MemDbURI)
	err := backupDB(dbDisk, db)

	db.QueryRow("select lastscan from meta").Scan(&lastscan)
}

func (h *Catalog) TrkExists(trk string) bool {
	var r int
	h.db.QueryRow("select count(trk) from tracks where trk = ?", trk).Scan(&r)
	if r == 1 {
		return true
	}
	return false
}
