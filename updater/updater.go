package updater

import (
	"fmt"

	"github.com/bogem/id3v2/v2"
	_ "github.com/mattn/go-sqlite3"
)

var id3opts id3v2.Options

func init() {
	id3opts = id3v2.Options{Parse: true}
}

func GetTag(f string) (*id3v2.Tag, error) {
	tag, err := id3v2.Open(f, id3opts)
	if err != nil {
		return nil, fmt.Errorf("'%s': %s", f, err)
	}
	tag.Close()
	return tag, err
}
