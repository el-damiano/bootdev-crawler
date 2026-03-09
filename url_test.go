package main

import (
	"fmt"
	"testing"
)

func TestURLnormalize(t *testing.T) {
	bootDevURLnormalized := "blog.boot.dev/path"
	cases := map[string]struct {
		want     string
		wantErr  bool
		inputURL string
	}{
		"already normalized": {
			inputURL: bootDevURLnormalized,
			want:     bootDevURLnormalized,
			wantErr:  false,
		},
		"https trailing slash": {
			inputURL: "https://blog.boot.dev/path/",
			want:     bootDevURLnormalized,
			wantErr:  false,
		},
		"https": {
			inputURL: "https://blog.boot.dev/path",
			want:     bootDevURLnormalized,
			wantErr:  false,
		},
		"http trailing slash": {
			inputURL: "http://blog.boot.dev/path/",
			want:     bootDevURLnormalized,
			wantErr:  false,
		},
		"http": {
			inputURL: "http://blog.boot.dev/path",
			want:     bootDevURLnormalized,
			wantErr:  false,
		},
		"https more subpaths": {
			inputURL: "https://blog.boot.dev/path/path2/path3/",
			want:     "blog.boot.dev/path/path2/path3",
			wantErr:  false,
		},
		"valid url with spaces and capitals": {
			inputURL: "HTTps://BLOG.boot.dev/PATH to somewhere",
			want:     "blog.boot.dev/path to somewhere",
			wantErr:  false,
		},
		"relative path without host": {
			inputURL: "/relative/path/without/host",
			want:     "/relative/path/without/host",
			wantErr:  false,
		},
		"empty input": {
			inputURL: "",
			want:     "",
			wantErr:  true,
		},
		"missing protocol scheme": {
			inputURL: "://invalidURL",
			want:     "",
			wantErr:  true,
		},
		"normal string of text": {
			inputURL: "you expected a valid url, but it was me, Dio!",
			want:     "you expected a valid url, but it was me, dio!",
			wantErr:  false,
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("Test case %v", i), func(t *testing.T) {
			got, err := urlNormalize(c.inputURL)
			if (err != nil) != c.wantErr {
				t.Errorf("FAIL: unexpected error: %v", err)
				return
			}
			if c.want != got {
				t.Errorf("FAIL: expected URL: %v, got: %v", c.want, got)
				return
			}
		})
	}
}
