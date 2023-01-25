package leases

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"regexp"
	"strings"
	"text/template"
	"time"
)

/*
Lease format specified in man(5) dhcpd.leases

- See https://linux.die.net/man/5/dhcpd.leases or similar

A Lease contains the data specified in the following format:

	172.24.43.3 {
		starts 6 2019/04/27 03:24:45;
		ends 6 2019/04/27 03:34:45;
		tstp 6 2019/04/27 03:34:45;
		cltt 6 2019/04/27 03:24:45;
		binding state free;
		hardware ethernet 00:db:70:c3:11:d7;
		uid "\001\000\333p\303\021\327";
	}
*/
type Lease struct {
	// IP address given to the lease
	IP net.IP `json:"ip"`

	// Start time of the lease
	Starts time.Time `json:"starts"`

	// Time when the lease expires
	Ends time.Time `json:"ends"`

	// Tstp is specified if the failover protocol is being used, and indicates what time the peer has been told the lease expires.
	Tstp time.Time `json:"tstp"`

	// Tsfp is also specified if the failover protocol is being used, and indicates the lease expiry time that the peer has acknowledged
	Tsfp time.Time `json:"tsfp"`

	// Atsfp is the actual time sent from the failover partner
	Atsfp time.Time `json:"atsfp,omitempty"`

	// Cltt is the client's last transaction time
	Cltt time.Time `json:"cllt"`

	/*The binding state statement declares the lease's binding state. When the DHCP server is
	not configured to use the failover protocol, a lease's binding state will be either
	active or free. The failover protocol adds some additional transitional states, as
	well as the backup state, which indicates that the lease is available for allocation
	by the failover secondary.*/
	BindingState string `json:"binding-state"`

	// The next binding state statement indicates what state the lease will move to when the current state expires. The time when the current state expires is specified in the ends statement.
	NextBindingState string `json:"next-binding-state"`

	/*Rewind binding state is used in failover. If the two servers go into communications-interrupted mode where they lose contact with each
	other, normally a particular server can only hand out new leases from	its share of the free pool. The idea is to allow it to reset the
	binding state of a lease so that it can re-issue the IP address*/
	RewindBindingState string `json:"rewind-binding-state"`

	// The hardware statement records the MAC address of the network interface on which the lease will be used. It is specified as a series of hexadecimal octets, separated by colons.
	Hardware Hardware `json:"hardware"`

	// The uid statement records the client identifier used by the client to acquire the lease. Clients are not required to send client identifiers, and this statement only appears if the client did in fact send one. Client identifiers are normally an ARP type (1 for ethernet) followed by the MAC address, just like in the hardware statement, but this is not required.
	UID string `json:"uid"`

	// Clients provided hostname
	ClientHostname string `json:"client-hostname"`

	// Optional settings for the lease, e.g., as the result of conditional evaluation performed by the server for the packet
	VendorClassID string `json:"vendor-class-identifier"`
	VendorName    string `json:"vendor-nmae"`

	// Optional circuit and remote ID suboption values if provided by the relay agent
	RelayCircuitId string
	RelayRemoteId  string
}

// Hardware is a representaion of the MAC and hardware types"
type Hardware struct {
	Hardware string           `json:"hardware"`
	MAC      string           `json:"mac"`
	MACAddr  net.HardwareAddr `json:"-"`
}

const leaseString = `
lease {{.IP}} {
  starts 4 {{.Starts.Format "2006/01/02 15:04:05"}};
  ends 4 {{.Ends.Format "2006/01/02 15:04:05"}};
  tstp 5 {{.Tstp.Format "2006/01/02 15:04:05"}};
  tsfp 6 {{.Tsfp.Format "2006/01/02 15:04:05"}};
  cltt 4 {{.Cltt.Format "2006/01/02 15:04:05"}};
  binding state {{.BindingState}};
  client-hostname "{{.ClientHostname}}";
  next binding state {{.NextBindingState}};
  hardware ethernet {{.Hardware.MAC}};
  uid "{{.UID}}";
{{- if not .Atsfp.IsZero}}
  cltt 4 {{.Atsfp.Format "2006/01/02 15:04:05"}};{{end}}
{{- if .RewindBindingState}}
  rewind binding state {{.RewindBindingState}};{{end}}
{{- if .VendorClassID}}
  set vendor-class-identifier = "{{.VendorClassID}}";{{end}}
{{- if .VendorName}}
  set vendor-name = "{{.VendorName}}";{{end}}
{{- if .RelayCircuitId}}
  option agent.circuit-id {{.RelayCircuitId}};{{end}}
{{- if .RelayRemoteId}}
  option agent.remote-id {{.RelayRemoteId}};{{end}}
}
`

var (
	// determine of time has a timezone suffix
	timeZoneRegex = regexp.MustCompile(`[a-zA-Z]{2,}$`)

	decoders = map[*regexp.Regexp]func(*Lease, string){
		regexp.MustCompile("lease ([\\d\\.]+) {"):                        func(l *Lease, line string) { l.IP = net.ParseIP(line) },
		regexp.MustCompile("host (.*) {"):                                func(l *Lease, line string) { l.ClientHostname = line },
		regexp.MustCompile("fixed-address (.*);"):                        func(l *Lease, line string) { l.IP = net.ParseIP(line) },
		regexp.MustCompile("cltt (?P<D>.*);"):                            func(l *Lease, line string) { l.Cltt = parseTime(line) },
		regexp.MustCompile("starts (?P<D>.*);"):                          func(l *Lease, line string) { l.Starts = parseTime(line) },
		regexp.MustCompile("ends (?P<D>.*);"):                            func(l *Lease, line string) { l.Ends = parseTime(line) },
		regexp.MustCompile("tsfp (?P<D>.*);"):                            func(l *Lease, line string) { l.Tsfp = parseTime(line) },
		regexp.MustCompile("tstp (?P<D>.*);"):                            func(l *Lease, line string) { l.Tstp = parseTime(line) },
		regexp.MustCompile("atsfp (?P<D>.*);"):                           func(l *Lease, line string) { l.Atsfp = parseTime(line) },
		regexp.MustCompile("cltt (?P<D>.*);"):                            func(l *Lease, line string) { l.Cltt = parseTime(line) },
		regexp.MustCompile(`uid "(?P<D>.*)";`):                           func(l *Lease, line string) { l.UID = line },
		regexp.MustCompile(`client-hostname "(?P<D>.*)";`):               func(l *Lease, line string) { l.ClientHostname = line },
		regexp.MustCompile(`(?m)^\s*binding state (?P<D>.*);`):           func(l *Lease, line string) { l.BindingState = line },
		regexp.MustCompile("next binding state (?P<D>.*);"):              func(l *Lease, line string) { l.NextBindingState = line },
		regexp.MustCompile("rewind binding state (?P<D>.*);"):            func(l *Lease, line string) { l.RewindBindingState = line },
		regexp.MustCompile(`option agent.circuit-id (?P<D>.*);`):         func(l *Lease, line string) { l.RelayCircuitId = line },
		regexp.MustCompile(`option agent.remote-id (?P<D>.*);`):          func(l *Lease, line string) { l.RelayRemoteId = line },
		regexp.MustCompile(`set vendor-class-identifier = "(?P<D>.*)";`): func(l *Lease, line string) { l.VendorClassID = line },
		regexp.MustCompile(`set vendor-name = "(?P<D>.*)";`):             func(l *Lease, line string) { l.VendorName = line },

		regexp.MustCompile("hardware (?P<D>.*);"): func(l *Lease, line string) {
			s := strings.SplitN(line, " ", 2)
			l.Hardware.Hardware = s[0]
			l.Hardware.MAC = s[1]
			if m, e := net.ParseMAC(s[1]); e == nil {
				l.Hardware.MACAddr = m
			}
		},
	}
)

/*parseTime from the off format of "6 2019/04/27 03:34:45;" and returns a time struct*/
func parseTime(s string) (t time.Time) {
	for _, fmt := range []string{
		"2006/01/02 15:04:05 +0000 MST", // with timezone
		"2006/01/02 15:04:05 MST",       // with timezone
		"2006/01/02 15:04:05",
	} {
		if t, err := time.Parse(fmt, s[2:]); err == nil {
			return t
		}
	}
	return time.Time{}
}

/*
parse takes a byte slice that looks like:

	172.24.43.3 {
		starts 6 2019/04/27 03:24:45;
		ends 6 2019/04/27 03:34:45;
		tstp 6 2019/04/27 03:34:45;
		cltt 6 2019/04/27 03:24:45;
		binding state free;
		hardware ethernet 00:db:70:c3:11:d7;
		uid "\001\000\333p\303\021\327";
	}

And populates the value of l with the values recoded
*/
func (l *Lease) parse(s []byte) {
	buf := bytes.NewBuffer(s)
	scanner := bufio.NewScanner(buf)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := scanner.Text()
		for re, fxn := range decoders {
			if re.MatchString(line) {
				s := re.FindStringSubmatch(line)
				fxn(l, s[1])
			}
		}
	}
}

func (l *Lease) String() string {
	var buf bytes.Buffer
	t := template.Must(template.New("lease").Parse(leaseString))
	if err := t.Execute(&buf, l); err != nil {
		fmt.Println("Failed executing template: ", err)
	}
	return buf.String()
}
