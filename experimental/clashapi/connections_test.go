package clashapi

import (
	"testing"

	"github.com/sagernet/sing-box/experimental/clashapi/trafficontrol"
)

func TestParseConnectionState(t *testing.T) {
	cases := []struct {
		raw  string
		want trafficontrol.ConnectionState
		ok   bool
	}{
		{raw: "", want: trafficontrol.ConnectionStateActive, ok: true},
		{raw: "active", want: trafficontrol.ConnectionStateActive, ok: true},
		{raw: "closed", want: trafficontrol.ConnectionStateClosed, ok: true},
		{raw: "all", want: trafficontrol.ConnectionStateAll, ok: true},
		{raw: "bad", want: trafficontrol.ConnectionStateActive, ok: false},
	}

	for _, tc := range cases {
		t.Run(tc.raw, func(t *testing.T) {
			got, ok := parseConnectionState(tc.raw)
			if ok != tc.ok || got != tc.want {
				t.Fatalf("parseConnectionState(%q) = %v, %v; want %v, %v", tc.raw, got, ok, tc.want, tc.ok)
			}
		})
	}
}
