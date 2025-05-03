package updater

import (
	//"fmt"

	"github.com/bogem/id3v2/v2"
	_ "github.com/mattn/go-sqlite3"
)

var id3opts id3v2.Options

func init() {
	id3opts = id3v2.Options{Parse: true}
}

