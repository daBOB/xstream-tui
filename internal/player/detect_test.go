package player

import (
	"testing"
)

func TestAvailability_Preferred(t *testing.T) {
	tests := []struct {
		name string
		a    Availability
		want PlayerType
	}{
		{"both available prefers mpv", Availability{MPV: true, VLC: true}, TypeMPV},
		{"only mpv", Availability{MPV: true, VLC: false}, TypeMPV},
		{"only vlc", Availability{MPV: false, VLC: true}, TypeVLC},
		{"none available", Availability{MPV: false, VLC: false}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.Preferred(); got != tt.want {
				t.Errorf("Preferred() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAvailability_String(t *testing.T) {
	tests := []struct {
		name string
		a    Availability
		want string
	}{
		{"both", Availability{MPV: true, VLC: true}, "mpv, vlc"},
		{"only mpv", Availability{MPV: true, VLC: false}, "mpv"},
		{"only vlc", Availability{MPV: false, VLC: true}, "vlc"},
		{"none", Availability{MPV: false, VLC: false}, "none"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDetect(t *testing.T) {
	// Just verify Detect() runs without panic and returns valid struct
	a := Detect()
	// We can't assume any player is installed, but we can verify the struct
	_ = a.MPV
	_ = a.VLC
	_ = a.Preferred()
	_ = a.String()
}
