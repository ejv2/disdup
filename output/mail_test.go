package output_test

import (
	"errors"
	"testing"

	"github.com/ejv2/disdup/output"
)

// Test port and address parsing.
func TestAddrParsing(t *testing.T) {
	cases := []struct {
		m          output.MailServer
		ExpectErr  error
		ExpectHost string
		ExpectPort int
	}{
		{output.MailServer{Address: "abcd:1234"}, nil, "abcd", 1234},
		{output.MailServer{Address: "abcd:-1234"}, output.ErrBadServer, "", 0},
		{output.MailServer{Address: "abcd"}, output.ErrBadServer, "", 0},
		{output.MailServer{Address: ":"}, output.ErrBadServer, "", 0},
		{output.MailServer{Address: ":1234"}, output.ErrBadServer, "", 0},
	}

	for _, c := range cases {
		gothost, gotport, err := c.m.AddrInfo()

		if c.ExpectErr != nil {
			if err == nil {
				t.Errorf("expected error from Open() (addr: %s), got nil", c.m.Address)
			} else if !errors.Is(err, c.ExpectErr) {
				t.Errorf("wrong error from Open()\nexpect: %s\ngot: %s", c.ExpectErr.Error(), err.Error())
			}

			continue
		} else if err != nil {
			t.Errorf("unexpected error from Open() (addr: %s)", c.m.Address)
			continue
		}

		if gothost != c.ExpectHost {
			t.Errorf("wrong host from Open()\nexpect: %s\ngot: %s", c.ExpectHost, gothost)
		}
		if gotport != c.ExpectPort {
			t.Errorf("wrong port from Open()\nexpect: %d\ngot: %d", c.ExpectPort, gotport)
		}
	}
}
