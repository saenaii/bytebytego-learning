package shortcode

import "testing"

func TestNormalize(t *testing.T) {
	tests := []struct {
		name      string
		raw       string
		mode      string
		blacklist []string
		want      string
		wantErr   bool
	}{
		{name: "sorts query and values", raw: "https://example.com/p?b=2&a=2&a=1", mode: ModeFull, want: "https://example.com/p?a=1&a=2&b=2"},
		{name: "removes blacklist", raw: "https://example.com/p?utm_source=x&b=2", mode: ModeFull, blacklist: []string{"utm_source"}, want: "https://example.com/p?b=2"},
		{name: "path only", raw: "https://example.com/p?a=1", mode: ModePathOnly, want: "https://example.com/p"},
		{name: "rejects relative", raw: "/p?a=1", mode: ModeFull, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Normalize(tt.raw, tt.mode, tt.blacklist)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Fatalf("Normalize() = %q, want %q", got, tt.want)
			}
		})
	}
}
