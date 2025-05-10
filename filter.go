package tmc

import (
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
	// chunks will be: &&, ||, ((, ))
	chunker = regexp.MustCompile(`&{2}|\|{2}|\({2}|\){2}`)
	// chunks will be: \\, //
	vchunker = regexp.MustCompile(`/{2}|\\{2}`)
	// chunks will be a SQL logical operator
	ychunker = regexp.MustCompile(`[<>=]{2}`)
}

func filterChunks(re *regexp.Regexp, input string) []string {
	binput := []byte(input)
	matches := re.FindAllIndex(binput, -1)
	if len(matches) == 0 {
		return []string{input}
	}

	c := []string{}
	if matches[0][0] != 0 {
		c = append(c, string(binput[:matches[0][0]]))
	}
	for i, m := range matches {
		// get the matching token
		c = append(c, string(binput[m[0]:m[1]]))
		// and the text following
		if i != len(matches) - 1 {
			// up to the start of the next match, if there
			// is a next match
			c = append(c, string(binput[m[1]:matches[i+1][0]]))
		} else {
			// and to end of slice if there isn't --
			// unless we're at the end of the slice
			if m[1] != len(binput) {
				c = append(c, string(binput[m[1]:]))
			}
		}
	}
	return c
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

	// do top-level chunking
	chunks := filterChunks(chunker, format)
	//fmt.Println(strings.Join(chunks, ";;"))

	// now parse chunks to build filter
	for _, chunk := range chunks {
		chunk = strings.TrimSpace(chunk)
		// handle logical operators, grouping, and empties
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
		// if the current attribute
		if attr == "facets" && !facets {
			open1 = fmt.Sprintf("%s, json_each(facets)", open1)
			open2 = fmt.Sprintf("%s, json_each(facets)", open2)
			facets = true
		}

		// split the value into chunks and iterate
		vchunks := filterChunks(vchunker, val)
		//fmt.Println(vchunks)
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
			ychunks := filterChunks(ychunker, vchunk)
			if len(ychunks) > 1 {
				filter = append(filter, fmt.Sprintf("%s ?", ychunks[0]))
				values = append(values, ychunks[1])
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
