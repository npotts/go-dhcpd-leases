
[![GoDoc](https://godoc.org/github.com/npotts/go-dhcpd-leases?status.svg)](https://godoc.org/github.com/npotts/go-dhcpd-leases)

# Go module to parse dhcpd's lease file

This little library alllows for parsing [dhcpd's lease files](https://linux.die.net/man/5/dhcpd.leases) as given in
`man 5 dhcpd.leases`.  Its not elegant, but gets the job done.

It does not parse all the options that dhcpd can emit, but it solves the problem I had.

# Usage

Basically do something like this:

```go
    f, err := os.Open("/var/lib/dhcpd/dhcpd.leases")
    leases := Parse(f)
    ...
```

