package leases

import (
	"bytes"
	"testing"
	"time"
)

func TestParseLease(t *testing.T) {
	in := []byte(`# The format of this file is documented in the dhcpd.leases(5) manual page.
# This lease file was written by isc-dhcp-4.3.6-P1

# authoring-byte-order entry is generated, DO NOT DELETE
authoring-byte-order little-endian;

lease 172.24.43.3 {
	starts 6 2019/04/27 03:24:45;
	ends 6 2019/04/27 03:34:45;
	tstp 6 2019/04/27 03:34:45;
	tsfp 6 2019/04/27 03:34:45;
	cltt 6 2019/04/27 03:24:45;
	atsfp 6 2019/04/27 03:34:45;
	client-hostname "gertrude";
	binding state active;
	next binding state free;
	hardware ethernet 01:34:56:67:89:9a;
	uid "\001\000\333p\303\021\327";
}
lease 172.24.43.4 {

`)

	buf := bytes.NewBuffer(in)
	i := Parse(buf)
	if i == nil {
		t.Errorf("Expect one lease")
	}
}

func TestParse(t *testing.T) {
	a := parseTime("6 2019/04/27 03:34:45")
	ex := time.Date(2019, 4, 27, 3, 34, 45, 0, time.UTC)

	if a.IsZero() {
		t.Error("Didnt parse time right")
	}
	if !a.Equal(ex) {
		t.Log("a ", a)
		t.Log("ex", ex)
		t.Error("Didnt parse time correctly")
	}
}
