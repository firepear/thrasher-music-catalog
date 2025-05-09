package tmc

import(
	"fmt"
	"regexp"
	"strings"
)

var (
	chunker  *regexp.Regexp
	vchunker *regexp.Regexp
	ychunker *regexp.Regexp
)

func init() {
	// chunks will be: &&, ||, ((, )), or anything else
	chunker = regexp.MustCompile(`(&&|\|\||\(\(|\)\)|.+)`)
	// chunks will be: \\, //, or anything else
	vchunker = regexp.MustCompile(`(//|\\\\|.+)`)
	// this one's easier to read
	ychunker = regexp.MustCompile(`^(?P<op>[<>=]+)(?P<val>[^<>=]+)$`)
}

// Filter takes a filter format string and turns it into a SQL
// statement and a list of values for that statement's
// placeholders. These, and the count of matching tracks, are stored
// in c.FltrStr, c.FltrVals, and c.FltrCount, respectively
func (c *Catalog) Filter(format  string) error {
	var err error
	var facets bool
	open1 := "SELECT trk FROM tracks"
	open2 := "SELECT count(trk) FROM tracks"
	filter := []string{"WHERE"}
	values := []any{}

	// do top-level chunking and iterate
	chunks := chunker.FindAllString(format, -1)
	//fmt.Println(strings.Join(chunks, ";;"))
	for _, chunk := range chunks {
		chunk = strings.TrimSpace(chunk)
		// handle logical operators
		if chunk == "||" {
			filter = append(filter, "OR")
			continue
		} else if chunk == "&&" {
			filter = append(filter, "AND")
			continue
		} else if chunk == "((" {
			filter = append(filter, "(")
			continue
		} else if chunk == "))" {
			filter = append(filter, ")")
			continue
		} else if chunk == "" {
			continue
		}

		// split the attribute and value
		attr, val, _ := strings.Cut(chunk, ":")
		if val == "" {
			return fmt.Errorf("attribute '%s' has no value", attr)
		}
		attr = strings.TrimSpace(attr)
		val = strings.TrimSpace(val)

		// normalize attributes
		attr, err = Normalize(attr)
		if err != nil {
			return err
		}
		if attr == "facets" && !facets {
			open1 = fmt.Sprintf("%s, json_each(facets)", open1)
			open2 = fmt.Sprintf("%s, json_each(facets)", open2)
			facets = true
		}

		// split the value into chunks and iterate
		vchunks := vchunker.FindAllString(val, -1)
		for _, vchunk := range vchunks {
			vchunk = strings.TrimSpace(vchunk)
			// handle logical ops, again
			if vchunk == "\\\\" {
				filter = append(filter, "AND")
				continue
			} else if vchunk == "//" {
				filter = append(filter, "OR")
				continue
			} else if vchunk == "" {
				continue
			}

			// now we have everything to turn this attr and value into SQL
			vchunk = strings.ReplaceAll(vchunk, "*", "%")
			if attr == "facets" {
				filter = append(filter, "json_each.value")
			} else {
				filter = append(filter, attr)
			}
			ychunks := ychunker.FindStringSubmatch(vchunk)
			if len(ychunks) > 0 {
				filter = append(filter, fmt.Sprintf("%s ?",
					ychunks[ychunker.SubexpIndex("op")]))
				values = append(values, ychunks[ychunker.SubexpIndex("val")])
			} else {
				filter = append(filter, "LIKE ?")
				values = append(values, vchunk)
			}
		}
	}

	// slap the count() opening clause onto the filter
	filter = append([]string{open2}, filter...)
	// run the query and and store the result in c.FltrCount
	err = c.db.QueryRow(strings.Join(filter, " "), values...).Scan(&c.FltrCount)
	// switch the count select for the regular one
	filter[0] = open1
	// store the finalized filter and its values
	c.FltrStr = strings.Join(filter, " ")
	c.FltrVals = values

	return err
}
