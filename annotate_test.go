package annotate

import (
	"bytes"
	"flag"
	"io"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"text/template"
)

var saveExp = flag.Bool("exp", false, "overwrite all expected output files with actual output (returning a failure)")
var match = flag.String("m", "", "only run tests whose name contains this string")

func TestWithHTML(t *testing.T) {
	annsByFile := map[string][]*Annotation{
		"hello_world.txt": []*Annotation{
			{0, 5, []byte("<b>"), []byte("</b>"), 0},
			{7, 12, []byte("<i>"), []byte("</i>"), 0},
		},
		"adjacent.txt": []*Annotation{
			{0, 3, []byte("<b>"), []byte("</b>"), 0},
			{3, 6, []byte("<i>"), []byte("</i>"), 0},
		},
		"empties.txt": []*Annotation{
			{0, 0, []byte("<b>"), []byte("</b>"), 0},
			{0, 0, []byte("<i>"), []byte("</i>"), 0},
			{2, 2, []byte("<i>"), []byte("</i>"), 0},
		},
		"nested_0.txt": []*Annotation{
			{0, 4, []byte("<1>"), []byte("</1>"), 0},
			{1, 3, []byte("<2>"), []byte("</2>"), 0},
		},
		"nested_1.txt": []*Annotation{
			{0, 4, []byte("<1>"), []byte("</1>"), 0},
			{1, 3, []byte("<2>"), []byte("</2>"), 0},
			{2, 2, []byte("<3>"), []byte("</3>"), 0},
		},
		"nested_2.txt": []*Annotation{
			{0, 2, []byte("<1>"), []byte("</1>"), 0},
			{2, 4, []byte("<2>"), []byte("</2>"), 0},
			{4, 6, []byte("<3>"), []byte("</3>"), 0},
			{7, 8, []byte("<4>"), []byte("</4>"), 0},
		},
		"html.txt": []*Annotation{
			{193, 203, []byte("<1>"), []byte("</1>"), 0},
			{336, 339, []byte("<WOOF>"), []byte("</WOOF>"), 0},
		},
	}

	dir := "testdata"
	tests, err := ioutil.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}

	for _, test := range tests {
		name := test.Name()
		if !strings.Contains(name, *match) {
			continue
		}
		if strings.HasSuffix(name, ".html") {
			continue
		}
		path := filepath.Join(dir, name)
		input, err := ioutil.ReadFile(path)
		if err != nil {
			t.Fatal(err)
			continue
		}

		anns := annsByFile[name]
		var buf bytes.Buffer
		err = WithHTML(input, anns, func(w io.Writer, b []byte) {
			template.HTMLEscape(w, b)
		}, &buf)
		if err != nil {
			t.Errorf("%s: WithHTML: %s", name, err)
			continue
		}
		got := buf.Bytes()

		expPath := path + ".html"
		if *saveExp {
			err = ioutil.WriteFile(expPath, got, 0700)
			if err != nil {
				t.Fatal(err)
			}
			continue
		}

		want, err := ioutil.ReadFile(expPath)
		if err != nil {
			t.Fatal(err)
		}

		want = bytes.TrimSpace(want)
		got = bytes.TrimSpace(got)

		if !bytes.Equal(want, got) {
			t.Errorf("%s: want %q, got %q", name, want, got)
			continue
		}
	}

	if *saveExp {
		t.Fatal("overwrote all expected output files with actual output (run tests again without -exp)")
	}
}

func BenchmarkAnnotate(b *testing.B) {
	input := []byte(strings.Repeat("a", 2000))
	n := len(input)/2 - 50
	anns := make([]*Annotation, n)
	for i := 0; i < n; i++ {
		if i%2 == 0 {
			anns[i] = &Annotation{Start: 2 * i, End: 2*i + 1}
		} else {
			anns[i] = &Annotation{Start: 2*i - 50, End: 2*i + 50}
			if anns[i].Start < 0 {
				anns[i].Start = 0
				anns[i].End = i
			}
			if anns[i].End >= len(input) {
				anns[i].End = len(input) - 1
			}
		}
		anns[i].WantInner = i % 5
	}
	sort.Sort(annotations(anns))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Annotate(input, 0, len(input), anns)
		if err != nil {
			b.Fatal(err)
		}
	}
}
