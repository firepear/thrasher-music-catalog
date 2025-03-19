package main

import (
	"database/sql"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

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

		err = filepath.WalkDir(music, func(path string, info fs.DirEntry, err error) error {
			if strings.HasSuffix(info.Name(), ".mp3") {
				tag, err := tmcu.GetTag(path)
				if err != nil {
					return err
				}
				fmt.Printf("%s | %s | %s | %s | %s\n", path,
					tag.Year(), tag.Artist(), tag.Album(), tag.Genre())
			}
			return err
		})
		if err != nil {
			fmt.Printf("error during scan: %s\n", err)
			os.Exit(3)
		}
	}
}
