package leases

import (
	"bytes"
	"net"
	"reflect"
	"testing"
	"time"
)

func TestParseLease(t *testing.T) {
	in := []byte(`# The format of this file is documented in the dhcpd.leases(5) manual page.
# This lease file was written by isc-dhcp-4.3.6-P1

# authoring-byte-order entry is generated, DO NOT DELETE
authoring-byte-order little-endian;

lease 172.24.43.3 {
	starts 6 2019/04/27 03:34:45;
	ends 6 2019/04/27 03:34:45;
	tstp 6 2019/04/27 03:34:45;
	tsfp 6 2019/04/27 03:34:45;
	cltt 6 2019/04/27 03:34:45;
	atsfp 6 2019/04/27 03:34:45;
	client-hostname "gertrude";
	binding state active;
	next binding state free;
	hardware ethernet 01:34:56:67:89:9a;
	uid "\001\000\333p\303\021\327";
}
lease 172.24.43.4 {
  starts 4 2020/04/27 03:34:45;
  ends 4 2020/04/27 03:34:45;
  tstp 5 2020/04/27 03:34:45;
  tsfp 6 2020/04/27 03:34:45;
  cltt 4 2020/04/27 03:34:45;
  binding state active;
  client-hostname "stein";
  next binding state expired;
  rewind binding state free;
  hardware ethernet 01:34:56:67:89:9b;
  uid "\001xr]?|d";
  set vendor-class-identifier = "maybe a switch";
  set vendor-name = "cisco something";
  option agent.circuit-id 0:4:3:ff:5:1b;
  option agent.remote-id 3:8:0:10:1:1:a:ad:80:a8;
}
`)

	buf := bytes.NewBuffer(in)
	ls := Parse(buf)
	if len(ls) != 2 {
		t.Errorf("Expect exactly two leases")
	}
	t1, _ := time.Parse("2006/01/02 15:04:05", "2019/04/27 03:34:45")
	hw1, _ := net.ParseMAC("01:34:56:67:89:9a")
	t2, _ := time.Parse("2006/01/02 15:04:05", "2020/04/27 03:34:45")
	hw2, _ := net.ParseMAC("01:34:56:67:89:9b")
	want := []Lease{
		{
			IP:                 net.ParseIP("172.24.43.3"),
			Starts:             t1,
			Ends:               t1,
			Tstp:               t1,
			Tsfp:               t1,
			Atsfp:              t1,
			Cltt:               t1,
			BindingState:       "active",
			NextBindingState:   "free",
			RewindBindingState: "",
			Hardware: Hardware{
				Hardware: "ethernet",
				MAC:      "01:34:56:67:89:9a",
				MACAddr:  hw1,
			},
			UID:            `\001\000\333p\303\021\327`,
			ClientHostname: "gertrude",
			VendorClassID:  "",
			VendorName:     "",
			RelayCircuitId: "",
			RelayRemoteId:  "",
		},
		{
			IP:                 net.ParseIP("172.24.43.4"),
			Starts:             t2,
			Ends:               t2,
			Tstp:               t2,
			Tsfp:               t2,
			Cltt:               t2,
			BindingState:       "active",
			NextBindingState:   "expired",
			RewindBindingState: "free",
			Hardware: Hardware{
				Hardware: "ethernet",
				MAC:      "01:34:56:67:89:9b",
				MACAddr:  hw2,
			},
			UID:            `\001xr]?|d`,
			ClientHostname: "stein",
			VendorClassID:  "maybe a switch",
			VendorName:     "cisco something",
			RelayCircuitId: "0:4:3:ff:5:1b",
			RelayRemoteId:  "3:8:0:10:1:1:a:ad:80:a8",
		},
	}
	if !reflect.DeepEqual(want, ls) {
		t.Fatalf("expected:\n%+v\ngot:\n%+v", want, ls)
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

func TestParseTime(t *testing.T) {
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

func TestParseTimeZ(t *testing.T) {
	a := parseTime("6 2019/04/27 03:34:45 MDT")
	mdt, err := time.LoadLocation("America/Denver")
	if err != nil {
		t.Errorf("excected to load America/Denve TZinfo : %v", err)
	}
	ex := time.Date(2019, 4, 27, 3, 34, 45, 0, mdt)

	if a.IsZero() {
		t.Error("Didnt parse time right")
	}
	if !a.Equal(ex) {
		t.Log("a ", a)
		t.Log("ex", ex)
		t.Error("Didnt parse time correctly")
	}
}
