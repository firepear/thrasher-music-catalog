package tmc

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"

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
			if v != "" {
				f = append(f, v)
			}
		}
	}
	return f, err
}

func getartists(db *sql.DB, cutoff int) ([]string, error) {
	a := []string{}
	r := ""
	c := ""
	rows, err := db.Query("SELECT artist, COUNT(artist) AS y FROM tracks WHERE artist <> '' GROUP BY artist HAVING y > ? ORDER BY artist COLLATE NOCASE", cutoff)
	if err != nil {
		return a, err
	}
	defer rows.Close()
	for rows.Next() {
		_ = rows.Scan(&r, &c)
		a = append(a, r)
	}
	return a, err
}

///////////////////////////////////////////////////////// exported funcs

// Normalize returns a normalized attribute value for any valid short
// attribute value. Unknown attributes return an error.
func Normalize(attr string) (string, error) {
	switch attr {
	case "a", "artist":
		attr = "artist"
	case "b", "album":
		attr = "album"
	case "f", "facet", "facets":
		attr = "facets"
	case "n", "num":
		attr = "tnum"
	case "t", "title":
		attr = "title"
	case "y", "year":
		attr = "year"
	default:
		return "", fmt.Errorf("unknown attribute '%s'", attr)
	}
	return attr, nil
}

// read config file, if it exists
func ReadConfig() (*Config, error) {
	confFile := os.Getenv("HOME") + "/.config/tmcrc"
	_, err := os.Stat(confFile)
	if err != nil {
		return nil, fmt.Errorf("%s doesn't exist", confFile)
	}
	// file exists; read in config
	conf := &Config{}
	confraw, _ := os.ReadFile(confFile)
	err = json.Unmarshal(confraw, conf)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse config %s: %s", confFile, err)
	}
	return conf, err
}


////////////////////////////////////////////////////////////////////////

type Catalog struct {
	db         *sql.DB
	Artists    []string
	Facets     []string
	FltrStr    string
	FltrVals   []any
	FltrCount  int
	QueryStr   string
	QueryVals  []any
	Lastscan   int
	TrackCount int
	TrimPrefix string
}

type Config struct {
	ArtistCutoff int    `json:"artist_cutoff"`
	DbFile       string `json:"dbfile"`
	MusicDir     string `json:"musicdir"`
}

type Track struct {
	Ctime  int
	Mtime  int
	Num    int
	Artist string
	Title  string
	Album  string
	Year   int
	Facets string
}

// New returns a Catalog instance which can be queried in various
// ways
func New(conf *Config, dbname string) (*Catalog, error) {
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
		err = memrestore(db, conf.DbFile)
		if err != nil {
			return nil, err
		}
	}

	// initialize Catalog
	c := &Catalog{db: db}
	db.QueryRow("SELECT lastscan FROM meta").Scan(&c.Lastscan)
	db.QueryRow("SELECT count(trk) FROM tracks").Scan(&c.TrackCount)
	c.Facets, err = getfacets(db)
	c.Artists, err = getartists(db, conf.ArtistCutoff)

	return c, err
}

// Query returns (a portion of) the filtered track set. Takes three
// arguments: a string of comma-separated attributes (in the same
// format as Filter) which will become the ORDER BY clause; then the
// limit and offset for the query.
func (c *Catalog) Query(orderby string, limit, offset int) ([]string, error) {
	if c.FltrStr == "" {
		return nil, fmt.Errorf("no filter is set")
	}
	if c.FltrCount > 0 && offset >= c.FltrCount {
		return nil, fmt.Errorf("offset %d >= filtered set of %d", offset, c.FltrCount)
	}
	c.QueryStr = c.FltrStr

	// handle ORDER BY if we've been given one
	if orderby != "" {
		c.QueryStr = fmt.Sprintf("%s ORDER BY ", c.QueryStr)
		for _, oattr := range strings.Split(orderby, ",") {
			oattr, err := Normalize(oattr)
			if err != nil {
				return nil, err
			}
			xtra := ""
			if oattr == "album" || oattr == "title" {
				xtra = xtra + " COLLATE NOCASE"
			}
			c.QueryStr = c.QueryStr + oattr + xtra + ","
			//qvals = append(qvals, oattr)
		}
	}
	c.QueryStr = strings.TrimRight(c.QueryStr, ",")
	c.QueryVals = c.FltrVals

	// limit and offset
	if limit > 0 {
		c.QueryStr = fmt.Sprintf("%s LIMIT ? OFFSET ?", c.QueryStr)
		c.QueryVals = append(c.FltrVals, limit, offset)
	}

	// run query
	rows, err := c.db.Query(c.QueryStr, c.QueryVals...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	trks := []string{}
	for rows.Next() {
		var t string
		_ = rows.Scan(&t)
		if c.TrimPrefix != "" {
			t = strings.TrimPrefix(t, c.TrimPrefix)
		}
		trks = append(trks, t)
	}

	return trks, err
}

// QueryRecent returns all tracks belonging to the 25 most recently
// added albums
func (c *Catalog) QueryRecent() ([]string, error) {
	c.QueryStr = "SELECT trk FROM tracks WHERE album IN (SELECT DISTINCT album FROM tracks ORDER BY ctime DESC LIMIT 25) ORDER BY ctime"
	//c.QueryStr = "SELECT trk FROM tracks WHERE album IN (SELECT DISTINCT album FROM (SELECT trk, album, ctime from tracks ORDER BY ctime DESC LIMIT 200)) ORDER BY ctime"
	rows, err := c.db.Query(c.QueryStr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	trks := []string{}
	for rows.Next() {
		var t string
		_ = rows.Scan(&t)
		if c.TrimPrefix != "" {
			t = strings.TrimPrefix(t, c.TrimPrefix)
		}
		trks = append(trks, t)
	}
	return trks, err
}

// TrkExists returns a boolean, based on whether a given path is known
// in the DB
func (c *Catalog) TrkExists(path string) bool {
	var r int
	if c.TrimPrefix != "" {
		path = c.TrimPrefix + path
	}
	c.db.QueryRow("select count(trk) from tracks where trk = ?", path).Scan(&r)
	if r == 1 {
		return true
	}
	return false
}

// TrkInfo returns the catalog data for a track
func (c *Catalog) TrkInfo(trk string) *Track {
	if c.TrimPrefix != "" {
		trk = c.TrimPrefix + trk
	}
	row := c.db.QueryRow(`select title, artist, album, year, tnum, facets
                                   from tracks where trk = ?`, trk)
	t := &Track{}
	err := row.Scan(&t.Title, &t.Artist, &t.Album, &t.Year, &t.Num, &t.Facets)
	if err != nil {
		fmt.Println(err)
	}
	return t
}

// Close closes the DB connection held by a Catalog
func (c *Catalog) Close() {
	c.db.Close()
}
