package leases

import (
	"net"
	"testing"
	"time"
)

func TestLease_String(t *testing.T) {
	ex := time.Date(2019, 4, 27, 3, 34, 45, 0, time.UTC)
	tests := []struct {
		name string
		l    Lease
		want string
	}{
		{"smoke test", Lease{
			IP:                 []byte{1, 2, 3, 4},
			Starts:             ex,
			Ends:               ex,
			Tstp:               ex,
			Tsfp:               ex,
			Atsfp:              ex,
			Cltt:               ex,
			BindingState:       "bidinging-state",
			NextBindingState:   "next-binding-state",
			RewindBindingState: "rewind-binding-state",
			Hardware: Hardware{
				Hardware: "hardware",
				MAC:      "01:23:45:67:89:0a",
				MACAddr:  net.HardwareAddr{0x00, 0x00, 0x5e, 0x00, 0x53, 0x01},
			},
			UID:            "uid",
			ClientHostname: "client-hostname",
			VendorClassID:  "vendor-class-id",
			VendorName:     "vendor-name",
			RelayCircuitId: "relay-circuit-id",
			RelayRemoteId:  "relay-remote-id",
		}, "\nlease 1.2.3.4 {\n  starts 4 2019/04/27 03:34:45;\n  ends 4 2019/04/27 03:34:45;\n  tstp 5 2019/04/27 03:34:45;\n  tsfp 6 2019/04/27 03:34:45;\n  cltt 4 2019/04/27 03:34:45;\n  binding state bidinging-state;\n  client-hostname \"client-hostname\";\n  next binding state next-binding-state;\n  hardware ethernet 01:23:45:67:89:0a;\n  uid \"uid\";\n  cltt 4 2019/04/27 03:34:45;\n  rewind binding state rewind-binding-state;\n  set vendor-class-identifier = \"vendor-class-id\";\n  set vendor-name = \"vendor-name\";\n  option agent.circuit-id relay-circuit-id;\n  option agent.remote-id relay-remote-id;\n}\n"},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.l.String(); got != tt.want {
				t.Errorf("Lease.String() = %q, want\n%q", got, tt.want)
			}
		})
	}
}
