package main

import (
	"fmt"
	"net/url"
	"reflect"
	"testing"
)

func TestExtractPageData(t *testing.T) {
	inputURL := "https://crawler-test.com"
	inputBody := `<html><body>
        <h1>Test Title</h1>
        <p>This is the first paragraph.</p>
        <a href="/link1">Link 1</a>
        <img src="/image1.jpg" alt="Image 1">
    </body></html>`

	actual := extractPageData(inputBody, inputURL)

	expected := PageData{
		URL:            "https://crawler-test.com",
		Heading:        "Test Title",
		FirstParagraph: "This is the first paragraph.",
		OutgoingLinks:  []string{"https://crawler-test.com/link1"},
		ImageUrls:      []string{"https://crawler-test.com/image1.jpg"},
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("expected %+v, got %+v", expected, actual)
	}
}

func TestH1Exraction(t *testing.T) {
	testCases := map[string]struct {
		input    string
		expected string
	}{
		"normal": {
			input:    "<html><body><h1>Normal Heading</h1></body></html>",
			expected: "Normal Heading",
		},
		"multiple headings": {
			input:    "<html><body><h1>First Heading</h1><h1>Second Heading</h1></body></html>",
			expected: "First Heading",
		},
		"no headings": {
			input:    "<html><body>Tough luck buddy</body></html>",
			expected: "",
		},
		"non-html text with heading": {
			input:    "for now this <h1>is fine</h1>",
			expected: "is fine",
		},
		"unclosed heading": {
			input:    "<html><body>normal text<h1>Start of the heading</body></html>",
			expected: "Start of the heading",
		},
		"empty string": {
			input:    "",
			expected: "",
		},
		// untested case: nested headings, because they're not allowed in the current HTML standard
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("Test case %v", i), func(t *testing.T) {
			got := getFirstH1FromHTML(testCase.input)
			if testCase.expected != got {
				t.Errorf("FAIL: expected: %v, got: %v", testCase.expected, got)
				return
			}
		})
	}

}

func TestFirstParagraphExtraction(t *testing.T) {
	testCases := map[string]struct {
		input    string
		expected string
	}{
		"normal": {
			input:    "<html><body><p>outside of main</p><main><p>inside of main</p></main></body></html>",
			expected: "inside of main",
		},
		"multiline paragraph": {
			input: `<html><body><p>outside of
main
</p>
<main><p>inside
of main
 </p></body></html>`,
			expected: "inside\nof main\n ",
		},
		"no paragraphs": {
			input:    "<html><body>Tough luck buddy</body></html>",
			expected: "",
		},
		"non-html text with paragraph": {
			input:    "for now this <p>is fine</p>",
			expected: "is fine",
		},
		"unclosed paragraph": {
			input:    "<html><body>normal text<p>Start of the paragraph</body></html>",
			expected: "Start of the paragraph",
		},
		"empty string": {
			input:    "",
			expected: "",
		},
		// untested case: nested headings, because they're not allowed in the current HTML standard
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("Test case %v", i), func(t *testing.T) {
			got := getFirstParagraphFromHTML(testCase.input)
			if testCase.expected != got {
				t.Errorf("FAIL: expected: %v, got: %v", testCase.expected, got)
				return
			}
		})
	}

}

func TestGetUrlsFromHTML(t *testing.T) {
	testCases := map[string]struct {
		inputURL  string
		inputBody string
		expected  []string
		wantErr   bool
	}{
		"normal": {
			inputURL:  "https://blog.boot.dev",
			inputBody: `<html><body><a href="https://blog.boot.dev"><span>Boot.dev</span></a></body></html>`,
			expected:  []string{"https://blog.boot.dev"},
			wantErr:   false,
		},
		"mulitple anchors": {
			inputURL:  "https://blog.boot.dev",
			inputBody: `<html><body><img src="/logo.png"><a href="https://blog.boot.dev"><a href="https://blog.boot.dev"><span>Boot.dev</span></a></body></html>`,
			expected:  []string{"https://blog.boot.dev", "https://blog.boot.dev"},
			wantErr:   false,
		},
		"no anchors or images": {
			inputURL:  "https://blog.boot.dev",
			inputBody: `<html><body></body></html>`,
			expected:  nil,
			wantErr:   false,
		},
		"<a> without href": {
			inputURL:  "https://blog.boot.dev",
			inputBody: `<html><body><a></a></body></html>`,
			expected:  nil,
			wantErr:   false,
		},
		"<a> with empty href": {
			inputURL:  "https://blog.boot.dev",
			inputBody: `<html><body><a href=""></a></body></html>`,
			expected:  nil,
			wantErr:   false,
		},
		"empty base url": {
			inputURL:  "",
			inputBody: `<html><body></body></html>`,
			expected:  nil,
			wantErr:   false,
		},
		"URL parsing error": {
			inputURL:  "",
			inputBody: `<html><body></body></html>`,
			expected:  nil,
			wantErr:   false,
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("Test case %v", i), func(t *testing.T) {
			baseURL, err := url.Parse(testCase.inputURL)
			if (err != nil) != testCase.wantErr {
				t.Errorf("FAIL: could not parse URL: %v", err)
				return
			}

			got, err := getUrlsFromHTML(testCase.inputBody, baseURL)
			if (err != nil) != testCase.wantErr {
				t.Errorf("FAIL: unexpected error: %v", err)
				return
			}

			if !reflect.DeepEqual(testCase.expected, got) {
				t.Errorf("FAIL: expected %v, got %v", testCase.expected, got)
			}
		})
	}
}

func TestGetImageUrlsFromHTML(t *testing.T) {
	testCases := map[string]struct {
		inputURL  string
		inputBody string
		expected  []string
		wantErr   bool
	}{
		"relative url from image": {
			inputURL:  "https://blog.boot.dev",
			inputBody: `<html><body><img src="/logo.png"></body></html>`,
			expected:  []string{"https://blog.boot.dev/logo.png"},
			wantErr:   false,
		},
		"<img> without src": {
			inputURL:  "https://blog.boot.dev",
			inputBody: `<html><body><img></body></html>`,
			expected:  nil,
			wantErr:   false,
		},
		"<img> with empty src": {
			inputURL:  "https://blog.boot.dev",
			inputBody: `<html><body><img src=""></body></html>`,
			expected:  nil,
			wantErr:   false,
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("Test case %v", i), func(t *testing.T) {
			baseURL, err := url.Parse(testCase.inputURL)
			if (err != nil) != testCase.wantErr {
				t.Errorf("FAIL: could not parse URL: %v", err)
				return
			}

			got, err := getImageUrlsFromHTML(testCase.inputBody, baseURL)
			if (err != nil) != testCase.wantErr {
				t.Errorf("FAIL: unexpected error: %v", err)
				return
			}

			if !reflect.DeepEqual(testCase.expected, got) {
				t.Errorf("FAIL: expected %v, got %v", testCase.expected, got)
			}
		})
	}
}
