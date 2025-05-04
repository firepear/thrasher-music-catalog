package updater

import (
	"database/sql"
	"fmt"
	//"os"

	"github.com/bogem/id3v2/v2"
	_ "github.com/mattn/go-sqlite3"
)

var id3opts id3v2.Options

func init() {
	id3opts = id3v2.Options{Parse: true}
}

type Updater struct {
	db *sql.DB
}

func New(dbfile string) (*Updater, error) {
	db, err := sql.Open("sqlite3", dbfile)
	if err != nil {
		return nil, err
	}
	return &Updater{db: db}, err
}

func (u *Updater) Close() {
	u.db.Close()
}

func (u *Updater) CreateDB() error {
	// test for emptiness
	var r int
	u.db.QueryRow("SELECT COUNT(name) FROM sqlite_master").Scan(&r)
	// if not zero, there's data in here
	if r > 0 {
		return fmt.Errorf("database not empty")
	}

	_, err := u.db.Exec(`CREATE TABLE tracks (
            trk TEXT UNIQUE,
            ctime INT,
            mtime INT,
            year INT,
            artist TEXT,
            album TEXT,
            title TEXT,
            tnum TEXT,
            facets TEXT)`)
	if err != nil {
		return err
	}

	_, err = u.db.Exec(`CREATE TABLE meta (
            lastscan int)`)
	if err != nil {
		return err
	}
	_, err = u.db.Exec(`INSERT INTO meta (lastscan) VALUES (0)`)
	return err
}

