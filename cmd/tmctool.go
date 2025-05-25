package main

import (
	"database/sql"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	//"time"

	tmc "github.com/firepear/thrasher-music-catalog"
	tmcu "github.com/firepear/thrasher-music-catalog/updater"

	_ "github.com/mattn/go-sqlite3"
)

var (
	cat      *tmc.Catalog
	fcreate  bool
	fscan    bool
	fadd     bool
	frm      bool
	fquery   bool
	fqquery  bool
	fqrecent bool
	fdebug   bool
	flimit   int
	foffset  int
	fdbfile  string
	fmusic   string
	ffilter  string
	forder   string
	ftrim    string
	genres   map[int]string
	genreg   *regexp.Regexp
	conf     *tmc.Config
)

func init() {
	// read config file, if it exists
	var err error
	conf, err = tmc.ReadConfig()
	if err != nil {
		fmt.Printf("error reading config: %s; continuing with null config...\n", err)
		conf = &tmc.Config{}
	}

	// handle flags
	flag.BoolVar(&fcreate, "c", false, "create new db")
	flag.BoolVar(&fscan, "s", false, "scan for new tracks")
	flag.BoolVar(&fdebug, "d", false, "print debug info")
	flag.BoolVar(&fadd, "a", false, "add facet to tracks")
	flag.BoolVar(&frm, "r", false, "remove facet from tracks")
	flag.BoolVar(&fquery, "q", false, "query and print track paths")
	flag.BoolVar(&fqquery, "qq", false, "query and print track details")
	flag.BoolVar(&fqrecent, "qr", false, "query and print recent track paths")
	flag.IntVar(&flimit, "l", 0, "query limit (default: size of filter set)")
	flag.IntVar(&foffset, "o", 0, "query offset (default: 0)")
	flag.StringVar(&fdbfile, "db", "", "database file to use")
	flag.StringVar(&fmusic, "m", "", "music directory to scan")
	flag.StringVar(&ffilter, "f", "", "filter format string to operate on")
	flag.StringVar(&forder, "ob", "", "comma-delineated list of attributes to order query by")
	flag.StringVar(&ftrim, "t", "", "prefix to remove from track paths")
	flag.Parse()

	// if fdbfile is set, override dbfile
	if fdbfile != "" {
		conf.DbFile = fdbfile
	}
	// ditto musicdir
	if fmusic != "" {
		conf.MusicDir = fmusic
	}
	// and if we still don't have a dbfile, bail
	if conf.DbFile == "" {
		fmt.Println("database file must be specified; see -h")
		os.Exit(1)
	}

	// setup genre stuff
	genreg = regexp.MustCompile("[0-9]+")
	genres = map[int]string{
		0: "Blues", 1: "Classic Rock", 2: "Country", 3: "Dance", 4: "Disco", 5: "Funk",
		6: "Grunge", 7: "Hip-Hop", 8: "Jazz", 9: "Metal", 10: "New Age", 11: "Oldies",
		12: "Other", 13: "Pop", 14: "R&B", 15: "Rap", 16: "Reggae", 17: "Rock",
		18: "Techno", 19: "Industrial", 20: "Alternative", 21: "Ska", 22: "Death Metal",
		23: "Pranks", 24: "Soundtrack", 25: "Euro-Techno", 26: "Ambient", 27: "Trip-Hop",
		28: "Vocal", 29: "Jazz+Funk", 30: "Fusion", 31: "Trance", 32: "Classical",
		33: "Instrumental", 34: "Acid", 35: "House", 36: "Game", 37: "Sound Clip", 38: "Gospel",
		39: "Noise", 40: "AlternRock", 41: "Bass", 42: "Soul", 43: "Punk", 44: "Space",
		45: "Meditative", 46: "Instrumental Pop", 47: "Instrumental Rock", 48: "Ethnic",
		49: "Gothic", 50: "Darkwave", 51: "Techno-Industrial", 52: "Electronic", 53: "Pop-Folk",
		54: "Eurodance", 55: "Dream", 56: "Southern Rock", 57: "Comedy", 58: "Cult",
		59: "Gangsta Rap", 60: "Top 40", 61: "Christian Rap", 62: "Pop / Funk", 63: "Jungle",
		64: "Native American", 65: "Cabaret", 66: "New Wave", 67: "Psychedelic", 68: "Rave",
		69: "Showtunes", 70: "Trailer", 71: "Lo-Fi", 72: "Tribal", 73: "Acid Punk",
		74: "Acid Jazz", 75: "Polka", 76: "Retro", 77: "Musical", 78: "Rock & Roll",
		79: "Hard Rock", 80: "Folk", 81: "Folk-Rock", 82: "National Folk", 83: "Swing",
		84: "Fast Fusion", 85: "Bebob", 86: "Latin", 87: "Revival", 88: "Celtic",
		89: "Bluegrass", 90: "Avantgarde", 91: "Gothic Rock", 92: "Progressive Rock",
		93: "Psychedelic Rock", 94: "Symphonic Rock", 95: "Slow Rock", 96: "Big Band",
		97: "Chorus", 98: "Easy Listening", 99: "Acoustic", 100: "Humour", 101: "Speech",
		102: "Chanson", 103: "Opera", 104: "Chamber Music", 105: "Sonata", 106: "Symphony",
		107: "Booty Bass", 108: "Primus", 109: "Porn Groove", 110: "Satire", 111: "Slow Jam",
		112: "Club", 113: "Tango", 114: "Samba", 115: "Folklore", 116: "Ballad",
		117: "Power Ballad", 118: "Rhythmic Soul", 119: "Freestyle", 120: "Duet",
		121: "Punk Rock", 122: "Drum Solo", 123: "A Cappella", 124: "Euro-House",
		125: "Dance Hall", 126: "Goa", 127: "Drum & Bass", 128: "Club-House", 129: "Hardcore",
		130: "Terror", 131: "Indie", 132: "BritPop", 133: "Negerpunk", 134: "Polsk Punk",
		135: "Beat", 136: "Christian Gangsta Rap", 137: "Heavy Metal", 138: "Black Metal",
		139: "Crossover", 140: "Contemporary Christian", 141: "Christian Rock",
		142: "Merengue", 143: "Salsa", 144: "Thrash Metal", 145: "Anime", 146: "JPop",
		147: "Synthpop", 148: "Abstract", 149: "Art Rock", 150: "Baroque", 151: "Bhangra",
		152: "Big Beat", 153: "Breakbeat", 154: "Chillout", 155: "Downtempo", 156: "Dub",
		157: "EBM", 158: "Eclectic", 159: "Electro", 160: "Electroclash", 161: "Emo",
		162: "Experimental", 163: "Garage", 164: "Global", 165: "IDM", 166: "Illbient",
		167: "Industro-Goth", 168: "Jam Band", 169: "Krautrock", 170: "Leftfield",
		171: "Lounge", 172: "Math Rock", 173: "New Romantic", 174: "Nu-Breakz",
		175: "Post-Punk", 176: "Post-Rock", 177: "Psytrance", 178: "Shoegaze",
		179: "Space Rock", 180: "Trop Rock", 181: "World Music", 182: "Neoclassical",
		183: "Audiobook", 184: "Audio Theatre", 185: "Neue Deutsche Welle",
		186: "Podcast", 187: "Indie Rock", 188: "G-Funk", 189: "Dubstep", 190: "Garage Rock",
		191: "Psybient",
	}
}

func scanmp3s(conf *tmc.Config) error {
	db, err := sql.Open("sqlite3", conf.DbFile)
	if err != nil {
		return err
	}
	defer db.Close()
	db.Exec("PRAGMA synchronous=0")

	var seen = 0
	var updated = 0
	var clean = false
	var genre = ""
	var ctime int64
	var mtime int64

	stmt, _ := db.Prepare("INSERT INTO tracks VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)")

	// add new tracks
	err = filepath.WalkDir(conf.MusicDir, func(path string, info fs.DirEntry, err error) error {
		// if looking at a dir check mtime and mark clean
		// unless it's newer than lastscan
		if info.IsDir() {
			stat, _ := info.Info()
			if stat.ModTime().Unix() <= int64(cat.Lastscan) {
				clean = true
			} else {
				clean = false
			}
			return nil
		}

		if strings.HasSuffix(info.Name(), ".mp3") {
			seen++

			// do nothing if our parent dir is clean
			if clean {
				return nil
			}

			// see if track is already in DB
			if cat.TrkExists(path) {
				// for now ignore it. maybe in the
				// future do some kind of update? but
				// also maybe we handle that in-DB
				return nil
			}

			// set create and modified time; ensure that
			// ctime is set to the lowest value for
			// ingestion purposes
			stat, _ := info.Info()
			ctime = int64(stat.Sys().(*syscall.Stat_t).Ctim.Sec)
			mtime = stat.ModTime().Unix()
			if ctime > mtime {
				ctime = mtime
			}

			// get tag data
			tag, err := tmc.ReadTag(path)
			if err != nil {
				return err
			}

			// munge genre, if it's numeric
			genid := string(genreg.Find([]byte(tag.Genre())))
			if len(genid) == 0 {
				genre = tag.Genre()
			} else {
				gi, _ := strconv.Atoi(genid)
				genre = genres[gi]
			}

			// get track number
			tnum := tag.GetTextFrame("TRCK").Text
			tnum = strings.Split(tnum, "/")[0]
			if tnum == "" {
				// no empty track numbers; they create
				// spurious errs later on
				tnum = "99"
			}

			// fixup year
			year := tag.Year()
			if year == "" {
				// no blank years
				year = "9999"
			}
			ychunks := strings.Split(year, "-")
			if len(ychunks) == 3 {
				// no ISO formatted datestamps
				year = ychunks[0]
			}

			fmt.Printf("+ %s '%s' '%s' (%s; %s; %s)\n",
				strings.TrimSpace(tag.Artist()), strings.TrimSpace(tag.Album()),
				strings.TrimSpace(tag.Title()), tnum, year, genre)
			_, err = stmt.Exec(path, ctime, mtime,
				tnum, strings.TrimSpace(tag.Artist()), strings.TrimSpace(tag.Title()),
				strings.TrimSpace(tag.Album()), year, fmt.Sprintf(`["%s"]`, genre))
			if err != nil {
				return err
			}
			updated++
		}
		return err
	})
	if err != nil {
		return err
	}

	fmt.Printf("Totals: seen %d, updated %d\n", seen, updated)
	_, err = db.Exec("UPDATE meta SET lastscan=?", mtime)
	return err
}

func main() {
	var err error
	if fdebug {
		fmt.Printf("DEBUG> DbFile: %s; MusicDir: %s\n", conf.DbFile, conf.MusicDir)
	}

	// past this point we might need an updater instance
	upd, err := tmcu.New(conf.DbFile)
	if err != nil {
		fmt.Printf("error creating updater: %s", err)
		os.Exit(1)
	}
	defer upd.Close()

	// we've been asked to create the db; do so and exit
	if fcreate {
		err := upd.CreateDB()
		if err != nil {
			fmt.Printf("couldn't create db: %s\n", err)
			os.Exit(2)
		}
		fmt.Printf("database initialized in %s\n", conf.DbFile)
		os.Exit(0)
	}

	// everything else needs a catalog instance, so make one
	cat, err = tmc.New(conf, "tmctool")
	cat.TrimPrefix = ftrim
	if err != nil {
		fmt.Printf("error creating catalog: %s", err)
		os.Exit(1)
	}
	defer cat.Close()

	// scan for new tracks and exit
	if fscan {
		stat, err := os.Stat(conf.MusicDir)
		if err != nil {
			fmt.Printf("can't access musicdir '%s': %s\n", conf.MusicDir, err)
			os.Exit(3)
		}
		if !stat.IsDir() {
			fmt.Printf("%s is not a directory\n", conf.MusicDir)
			os.Exit(3)
		}

		err = scanmp3s(conf)
		if err != nil {
			fmt.Printf("error during scan: %s\n", err)
			os.Exit(3)
		}
		os.Exit(0)
	}

	if fqrecent == true {
		trks, err := cat.QueryRecent()
		if err != nil {
			fmt.Printf("error getting recent tracks: %s\n", err)
			os.Exit(3)
		}
		for _, trk := range trks {
			fmt.Println(trk)
		}
		os.Exit(0)
	}

	// handle setting filter, if we have a format string. bail if
	// we don't, because anything else requires that to be set
	if ffilter != "" {
		err = cat.Filter(ffilter)
		if err != nil {
			fmt.Printf("error parsing filter: %s\n", err)
			os.Exit(3)
		}
		if fdebug {
			fmt.Printf("DEBUG> filter: '%s', %v, %d\n", cat.FltrStr, cat.FltrVals, cat.FltrCount)
		}
	} else {
		fmt.Println("no op requested, or op requires a filter to be set; see the README")
		os.Exit(1)
	}
	trks := []string{}

	// query catalog and produce output
	if fquery || fqquery {
		trks, err = cat.Query(forder, flimit, foffset)
		if err != nil {
			fmt.Printf("error querying catalog: %s\n", err)
			os.Exit(2)
		}
		if fdebug {
			fmt.Printf("DEBUG> query: '%s', %v\n----\n", cat.QueryStr, cat.QueryVals)
		}
	}
	if fqquery {
		// fetch and print track details
		for _, trk := range trks {
			i := cat.TrkInfo(trk)
			if len(i.Artist) > 30 {
				i.Artist = i.Artist[:29] + "…"
			}
			if len(i.Title) > 50 {
				i.Title = i.Title[:49] + "…"
			}
			if len(i.Album) > 30 {
				i.Album = i.Album[:29] + "…"
			}
			fmt.Printf("%3d | %-30s | %-50s | %-30s | %d |\t%s\n",
				i.Num, i.Artist, i.Title, i.Album, i.Year, i.Facets)
		}
	} else {
		// just print the track paths
		for _, trk := range trks {
			fmt.Println(trk)
		}
	}
}
