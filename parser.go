package leases

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
)

var (
	lkwd       = []byte("\nlease ")
	hkwd       = []byte("\nhost ")
	closeParen = []byte("}")
)

/*Parse reads from a dhcpd.leases file and returns a list of leases.  Unknown fields are ignored
 */
func Parse(r io.Reader) []Lease {
	toLease := func(d []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF {
			return 0, nil, fmt.Errorf("Unable to parse")
		}

		i := bytes.Index(d, lkwd)
		if i == -1 {
			i = bytes.Index(d, hkwd)
		}

		if i != -1 { //locate folloing }
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
	return rtn
}
