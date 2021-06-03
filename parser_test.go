package leases

import (
	"bytes"
	"fmt"
	"net"
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

	for _, ii := range i {
		fmt.Println("ip: ", ii.IP)
	}
}

func TestParseHost(t *testing.T) {
	in := []byte(`# The format of this file is documented in the dhcpd.leases(5) manual page.
# This lease file was written by isc-dhcp-4.2.5

host test1.example.com {
  dynamic;
  hardware ethernet 4b:54:ef:7d:c3:0d;
  fixed-address 10.113.10.24;
        supersede server.filename = "pxelinux.0";
        supersede host-name = "test1.example.com";
}
host test2.example.com {
  dynamic;
  hardware ethernet c5:ea:cf:1e:2f:c9;
  fixed-address 10.113.10.9;
        supersede server.filename = "pxelinux.0";
        supersede host-name = "test2.example.com";
}
	`)

	buf := bytes.NewBuffer(in)
	i := Parse(buf)
	if i == nil {
		t.Errorf("Expect one lease")
	}
	if i[0].ClientHostname != "test1.example.com" {
		t.Errorf("Invalid hostname")
	}
	if !net.IPv4(10, 113, 10, 24).Equal(i[0].IP) {
		t.Errorf("Invalid IP: got %s", i[0].IP)
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
