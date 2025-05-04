package updater

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/bogem/id3v2/v2"
	_ "github.com/mattn/go-sqlite3"
)

var id3opts id3v2.Options

func init() {
	id3opts = id3v2.Options{Parse: true}
}

func CreateDB(dbfile string) error {
	_, err := os.Stat(dbfile)
	if err == nil {
		return fmt.Errorf("%s exists", dbfile)
	}

	db, err := sql.Open("sqlite3", dbfile)
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE tracks (
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

	_, err = db.Exec(`CREATE TABLE meta (
                            lastscan int)`)
	if err != nil {
		return err
	}
	_, err = db.Exec(`INSERT INTO meta (lastscan) VALUES (0)`)
	return err
}

