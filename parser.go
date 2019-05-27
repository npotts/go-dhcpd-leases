package leases

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
)

var (
	lkwd       = []byte("\nlease ")
	closeParen = []byte("}")
)

/*Parse reads from a dhcpd.leases file and returns a list of leases.  Unknown fields are ignored
 */
func Parse(r io.Reader) ([]Lease, error) {
	toLease := func(d []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF {
			return 0, nil, fmt.Errorf("Unable to parse")
		}
		if i := bytes.Index(d, lkwd); i != -1 { //locate folloing }
			i += 7
			if j := bytes.Index(d[i:], closeParen); j != -1 {
				return i + j + 1, d[i : i+j+1], nil
			}
		}
		return 0, nil, nil
	}

	scanner := bufio.NewScanner(r)
	scanner.Split(toLease)

	rtn := []Lease{}

	for scanner.Scan() {
		l := Lease{}
		l.parse(scanner.Bytes())
		rtn = append(rtn, l)

	}
	return rtn, nil
}
