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
	chunker = regexp.MustCompile(`(&&|\|\||\(\(|\)\)|[^&\|\(\)]+)`)
	// chunks will be: \\, //, or anything else
	vchunker = regexp.MustCompile(`(//|\\\\|[^/\\]+)`)
	// this one's easier to read
	ychunker = regexp.MustCompile(`([<>=]+|[0-9%]+)`)
}

// ParseFilter takes a filter format string and turns it into a SQL
// statement and a list of values for that statement's placeholders
func ParseFilter(c *Catalog, in string) (string, []string, error) {
	var err error
	values := []string{}
	clauses := []string{"SELECT trk FROM tracks WHERE"}

	// do top-level chunking and iterate
	chunks := chunker.FindAllString(in, -1)
	//fmt.Println(strings.Join(chunks, ";;"))
	for _, chunk := range chunks {
		chunk = strings.TrimSpace(chunk)
		// handle logical operators
		if chunk == "||" {
			clauses = append(clauses, "OR")
			continue
		} else if chunk == "&&" {
			clauses = append(clauses, "AND")
			continue
		} else if chunk == "((" {
			clauses = append(clauses, "(")
			continue
		} else if chunk == "))" {
			clauses = append(clauses, ")")
			continue
		} else if chunk == "" {
			continue
		}

		// split the attribute and value
		attr, val, _ := strings.Cut(chunk, ":")
		if val == "" {
			return "", values, fmt.Errorf("attribute '%s' has no value", attr)
		}
		attr = strings.TrimSpace(attr)
		val = strings.TrimSpace(val)

		// normalize attributes
		switch attr {
		case "a", "artist":
			attr = "artist"
		case "b", "album":
			attr = "album"
		case "t", "title":
			attr = "title"
		case "f", "facet", "facets":
			attr = "facets"
		case "y", "year":
			attr = "year"
		default:
			return "", values, fmt.Errorf("unknown attribute '%s'", attr)
		}

		// split the value into chunks and iterate
		vchunks := vchunker.FindAllString(val, -1)
		for _, vchunk := range vchunks {
			vchunk = strings.TrimSpace(vchunk)
			// handle logical ops, again
			if vchunk == "\\\\" {
				clauses = append(clauses, "AND")
				continue
			} else if vchunk == "//" {
				clauses = append(clauses, "OR")
				continue
			} else if vchunk == "" {
				continue
			}

			// now we have everything to turn this attr and value into SQL
			clauses = append(clauses, attr)
			if attr == "year" {
				ychunks := ychunker.FindAllString(vchunk, -1)
				if len(ychunks) == 2 {
					clauses = append(clauses, fmt.Sprintf("%s ?", ychunks[0]))
					values = append(values, ychunks[1])
				} else {
					clauses = append(clauses, "LIKE ?")
					values = append(values, vchunk)
				}
			} else if attr == "facets" {
				clauses = append(clauses, "LIKE ?")
				values = append(values, fmt.Sprintf("%%%s%%", vchunk))
			} else {
				clauses = append(clauses, "LIKE ?")
				values = append(values, vchunk)
			}
		}
	}

	return strings.Join(clauses, " "), values, err
}
