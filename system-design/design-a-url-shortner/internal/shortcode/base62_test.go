package shortcode

import "testing"

func TestEncode(t *testing.T) {
	tests := []struct {
		name string
		id   int64
		want string
	}{
		{name: "zero", id: 0, want: "0"},
		{name: "one", id: 1, want: "1"},
		{name: "alphabet boundary", id: 62, want: "10"},
		{name: "large id", id: 2009215674938, want: "zn9edcu"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Encode(tt.id)
			if err != nil || got != tt.want {
				t.Fatalf("Encode(%d) = %q, %v; want %q", tt.id, got, err, tt.want)
			}
		})
	}
}
