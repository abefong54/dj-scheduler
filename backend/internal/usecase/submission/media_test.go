package submission

import "testing"

func TestResolveMedia(t *testing.T) {
	cases := []struct {
		name         string
		url          string
		wantProvider string
		wantRef      string
		wantErr      bool
	}{
		{"youtube watch", "https://www.youtube.com/watch?v=dQw4w9WgXcQ", "youtube", "dQw4w9WgXcQ", false},
		{"youtube watch extra params", "https://youtube.com/watch?v=abc123XYZ&t=42s", "youtube", "abc123XYZ", false},
		{"youtube short domain", "https://youtu.be/dQw4w9WgXcQ", "youtube", "dQw4w9WgXcQ", false},
		{"youtube embed path", "https://www.youtube.com/embed/dQw4w9WgXcQ", "youtube", "dQw4w9WgXcQ", false},
		{"youtube shorts", "https://youtube.com/shorts/xyz789", "youtube", "xyz789", false},
		{"youtube music host", "https://music.youtube.com/watch?v=trackid00", "youtube", "trackid00", false},
		{"youtube no id", "https://www.youtube.com/", "", "", true},
		{"soundcloud track", "https://soundcloud.com/artist/my-mix", "soundcloud", "https://soundcloud.com/artist/my-mix", false},
		{"soundcloud www", "https://www.soundcloud.com/artist/set/my-set", "soundcloud", "https://soundcloud.com/artist/set/my-set", false},
		{"soundcloud bare", "https://soundcloud.com/", "", "", true},
		{"mixcloud show", "https://www.mixcloud.com/artist/my-show/", "mixcloud", "https://www.mixcloud.com/artist/my-show/", false},
		{"mixcloud bare", "https://mixcloud.com", "", "", true},
		{"unsupported host", "https://vimeo.com/12345", "", "", true},
		{"spotify rejected", "https://open.spotify.com/track/abc", "", "", true},
		{"not a url", "just some text", "", "", true},
		{"ftp scheme", "ftp://youtube.com/watch?v=abc", "", "", true},
		{"empty", "", "", "", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m, err := ResolveMedia(tc.url)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q, got %+v", tc.url, m)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for %q: %v", tc.url, err)
			}
			if m.Provider != tc.wantProvider || m.Ref != tc.wantRef {
				t.Fatalf("ResolveMedia(%q) = {%q,%q}, want {%q,%q}", tc.url, m.Provider, m.Ref, tc.wantProvider, tc.wantRef)
			}
		})
	}
}
