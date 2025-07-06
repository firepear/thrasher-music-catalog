package updater

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

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

func (u *Updater) SetFacets(trk string, facets []string) error {
	// prepare a statement
	stmt, _ := u.db.Prepare("UPDATE tracks SET facets = ?, mtime = ? WHERE trk = ?")
	var jfacets []byte
	var err error
	// stringify our list of facets
	if jfacets, err = json.Marshal(facets); err != nil {
		return err
	}
	// execute update
	_, err = stmt.Exec(string(jfacets), time.Now().Unix(), trk)
	return err
}

func (u *Updater) SetAlbum(trk, alb string) error {
	// prepare a statement
	stmt, _ := u.db.Prepare("UPDATE tracks SET album = ?, mtime = ? WHERE trk = ?")
	// execute update
	_, err := stmt.Exec(alb, time.Now().Unix(), trk)
	return err
}

func (u *Updater) SetArtist(trk, art string) error {
	// prepare a statement
	stmt, _ := u.db.Prepare("UPDATE tracks SET artist = ?, mtime = ? WHERE trk = ?")
	// execute update
	_, err := stmt.Exec(art, time.Now().Unix(), trk)
	return err
}

func (u *Updater) SetYear(trk string, year int) error {
	// prepare a statement
	stmt, _ := u.db.Prepare("UPDATE tracks SET year = ?, mtime = ? WHERE trk = ?")
	// execute update
	_, err := stmt.Exec(year, time.Now().Unix(), trk)
	return err
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
            tnum INT,
            artist TEXT,
            title TEXT,
            album TEXT,
            year INT,
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

