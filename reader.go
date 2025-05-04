package tmc

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/bogem/id3v2/v2"

	sqlite "github.com/mattn/go-sqlite3"
)

var id3opts id3v2.Options

func init() {
	id3opts = id3v2.Options{Parse: true}
}

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
			diskSqliteConn, ok := diskRawConn.(*sqlite.SQLiteConn)
			if !ok {
				return fmt.Errorf("error casting disk raw conn to sqlite conn")
			}
			memSqliteConn, ok := memRawConn.(*sqlite.SQLiteConn)
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

////////////////////////////////////////////////////// internal db funcs

func getfacets(db *sql.DB) ([]string, error) {
	f := []string{}
	fJson := []string{}

	rows, err := db.Query("SELECT facets FROM tracks GROUP BY facets")
	if err != nil {
		return f, err
	}
	defer rows.Close()

	for rows.Next() {
		var fRaw string
		_ = rows.Scan(&fRaw)
		_ = json.Unmarshal([]byte(fRaw), &fJson)
		for _, v := range fJson {
			if slices.Contains(f, v) {
				continue
			}
			f = append(f, v)
		}
	}

	return f, err
}

///////////////////////////////////////////////////////// exported funcs

// ReadTag takes a file path and returns the ID3 tags contained in
// that file
func ReadTag(path string) (*id3v2.Tag, error) {
	tag, err := id3v2.Open(path, id3opts)
	if err != nil {
		return nil, fmt.Errorf("'%s': %s", path, err)
	}
	tag.Close()
	return tag, err
}

////////////////////////////////////////////////////////////////////////

type Catalog struct {
	db        *sql.DB
	Facets    []string
	Filter    string
	FltrVals  []any
	FltrCount int
	Lastscan  int
}

type Track struct {
	Title  string
	Artist string
	Album  string
	Year   string
	Tnum   string
	Facets string
}

// New returns a Catalog instance which can be queried in various
// ways
func New(dbfile, dbname string) (*Catalog, error) {
	// open in-mem db in shared mode
	db, err := sql.Open("sqlite3", fmt.Sprintf("file:%s?mode=memory&cache=shared", dbname))
	if err != nil {
		return nil, err
	}

	// test for emptiness
	var r int
	db.QueryRow("SELECT COUNT(name) FROM sqlite_master").Scan(&r)
	// if we got zero, we need to restore from disk
	if r == 0 {
		err = memrestore(db, dbfile)
		if err != nil {
			return nil, err
		}
	}

	// initialize Catalog
	c := &Catalog{db: db}
	db.QueryRow("SELECT lastscan FROM meta").Scan(&c.Lastscan)
	c.Facets, err = getfacets(db)

	return c, err
}

// Query returns (a portion of) the filtered track set. Takes two
// arguments, the limit and offset for the query.
func (c *Catalog) Query(limit, offset int) ([]string, error) {
	if offset >= c.FltrCount {
		return nil, fmt.Errorf("offset %d >= filtered set of %d", offset, c.FltrCount)
	}

	qstr := fmt.Sprintf("%s LIMIT ? OFFSET ?", c.Filter)
	rows, err := c.db.Query(qstr, append(c.FltrVals, limit, offset)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	trks := []string{}
	for rows.Next() {
		var t string
		_ = rows.Scan(&t)
		trks = append(trks, t)
	}
	return trks, err
}

// TrkExists returns a boolean, based on whether a given path is known
// in the DB
func (c *Catalog) TrkExists(path string) bool {
	var r int
	c.db.QueryRow("select count(trk) from tracks where trk = ?", path).Scan(&r)
	if r == 1 {
		return true
	}
	return false
}

// TrkInfo returns the catalog data for a track
func (c *Catalog) TrkInfo(trk string) *Track {
	row := c.db.QueryRow(`select title, artist, album, year, tnum, facets
                                   from tracks where trk = ?`, trk)
	t := &Track{}
	row.Scan(&t.Title, &t.Artist, &t.Album, &t.Year, &t.Tnum, &t.Facets)
	return t
}

// Close closes the DB connection held by a Catalog
func (c *Catalog) Close() {
	c.db.Close()
}
