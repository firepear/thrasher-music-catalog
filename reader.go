package tmc

import (
	"context"
	"database/sql"
	"fmt"

	sqlite "github.com/mattn/go-sqlite3"
)

////////////////////////////////////////////////////////// restore funcs
// the entirety of the restore code is taken from examples on the
// internet. it's in multiple places, posted by multiple people. seems
// there's basically one way to do this.
func memrestore(memdb *sql.DB, ddb string) error {
	// open diskdb
	diskdb, err := sql.Open("sqlite3", ddb)
	if err != nil {
		return err
	}
	defer diskdb.Close()

	// get Conns to memdb, diskdb
	memconn, err := memdb.Conn(context.Background())
	if err != nil {
		return err
	}
	diskconn, err := diskdb.Conn(context.Background())
	if err != nil {
		return err
	}

	// get the underlying sqlite connections
	err = diskconn.Raw(func(diskRawConn any) error {
		return memconn.Raw(func(memRawConn any) error {
			diskSqliteConn, ok :=  diskRawConn.(*sqlite.SQLiteConn)
			if !ok {
				return fmt.Errorf("error casting disk raw conn to sqlite conn")
			}
			memSqliteConn, ok :=  memRawConn.(*sqlite.SQLiteConn)
			if !ok {
				return fmt.Errorf("error casting mem raw conn to sqlite conn")
			}
			// start the actual restore with the sqlite connections
			return innerrestore(diskSqliteConn, memSqliteConn)
		})
	})
	return err
}

func innerrestore(diskConn, memConn *sqlite.SQLiteConn) error {
	b, err := memConn.Backup("main", diskConn, "main")
	if err != nil {
		return fmt.Errorf("error initializing SQLite backup: %w", err)
	}
	done, err := b.Step(-1)
	if !done {
		// it should never happen when using -1 as step
		return fmt.Errorf("generic error: backup is not done after step")
	}
	if err != nil {
		return fmt.Errorf("error in stepping backup: %w", err)
	}
	// remember to call finish to clear up resources
	err = b.Finish()
	if err != nil {
		return fmt.Errorf("error finishing backup: %w", err)
	}
	return nil
}
////////////////////////////////////////////////////// end restore funcs

type Catalog struct {
	db       *sql.DB
	Lastscan int
}

func New(diskdb string) (*Catalog, error) {
	// open in-mem db in shared mode
	db, err := sql.Open("sqlite3", "file:tmcdb?mode=memory&cache=shared")
	if err != nil {
		return nil, err
	}

	// test for emptiness
	var r int
	db.QueryRow("select count(name) from sqlite_master").Scan(&r)
	// if we got zero, we need to restore from disk
	if r == 0 {
		err = memrestore(db, diskdb)
		if err != nil {
			return nil, err
		}
	}

	c := &Catalog{db: db}
	db.QueryRow("select lastscan from meta").Scan(c.Lastscan)
	return c, err
}

func (h *Catalog) TrkExists(trk string) bool {
	var r int
	h.db.QueryRow("select count(trk) from tracks where trk = ?", trk).Scan(&r)
	if r == 1 {
		return true
	}
	return false
}
