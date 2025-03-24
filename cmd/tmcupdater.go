package main

import (
	"database/sql"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	//tmcq "github.com/firepear/thrasher-music-catalog"
	tmcu "github.com/firepear/thrasher-music-catalog/updater"

	_ "github.com/mattn/go-sqlite3"
)

var (
	fcreate bool
	fscan   bool
	fadd    bool
	frm     bool
	fdbfile string
)

func init() {
	flag.BoolVar(&fcreate, "c", false, "create new db")
	flag.BoolVar(&fscan, "s", false, "scan for new tracks")
	flag.BoolVar(&fadd, "a", false, "add facet to tracks")
	flag.BoolVar(&frm, "r", false, "remove facet from tracks")
	flag.StringVar(&fdbfile, "d", "", "database file to use")
	flag.Parse()
}

func createDB(dbfile string) error {
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
                            facets TEXT)`)
	_, err = db.Exec(`CREATE TABLE meta (
                            lastscan int)`)
	return err
}

func scanmp3s(musicdir, dbfile string) error {
	db, err := sql.Open("sqlite3", dbfile)
	if err != nil {
		return err
	}
	defer db.Close()
	db.Exec("PRAGMA synchronous=0")

	var lastscan = 0
	var seen     = 0
	var updated  = 0
	var clean    = false

	ctime   := time.Now().Unix()
	mtime   := ctime
	stmt, _ := db.Prepare("INSERT INTO tracks VALUES (?, ?, ?, ?, ?, ?, ?, ?)")

	// add new tracks
	err = filepath.WalkDir(musicdir, func(path string, info fs.DirEntry, err error) error {
		// if looking at a dir check mtime and mark clean
		// unless it's newer than lastscan
		if info.IsDir() {
			stat, _ := info.Info()
			if stat.ModTime().Unix() <= int64(lastscan) {
				clean = true
			} else {
				clean = false
			}
			fmt.Printf("s:%d, u:%d\n", seen, updated)
			return nil
		}

		if strings.HasSuffix(info.Name(), ".mp3") {
			// do nothing if our parent dir is clean
			seen++
			if clean {
				return nil
			}

			tag, err := tmcu.GetTag(path)
			if err != nil {
				return err
			}
			_, err = stmt.Exec(path, ctime, mtime,
				tag.Year(), tag.Artist(), tag.Album(), tag.Title(),
				fmt.Sprintf(`{"f":["%s"]}`, tag.Genre()))
			if err != nil {
				return err
			}
			//fmt.Printf("%s | %s | %s | %s | %s\n", path,
			//	tag.Year(), tag.Artist(), tag.Album(), tag.Genre())
			updated++
		}
		return err
	})
	db.Exec("UPDATE meta SET lastscan = ?", mtime)
	return err
}

func main() {
	// handle flags
	if fdbfile == "" {
		fmt.Println("database file must be specified; see -h")
		os.Exit(1)
	}

	if fcreate {
		err := createDB(fdbfile)
		if err != nil {
			fmt.Printf("couldn't create db: %s\n", err)
			os.Exit(2)
		}
		fmt.Printf("database initialized in %s\n", fdbfile)
	}
	if fscan {
		music := flag.Arg(0)
		stat, err := os.Stat(flag.Arg(0))
		if err != nil {
			fmt.Printf("can't access '%s': %s\n", music, err)
			os.Exit(3)
		}
		if !stat.IsDir() {
			fmt.Printf("%s is not a directory\n", music)
			os.Exit(3)
		}

		err = scanmp3s(music, fdbfile)
		if err != nil {
			fmt.Printf("error during scan: %s\n", err)
			os.Exit(3)
		}
	}
}
