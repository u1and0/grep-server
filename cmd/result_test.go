package cmd

import (
	"fmt"
	"testing"
)

func TestResult_highlightFilename(t *testing.T) {
	r := Result{} // Normal Case
	s := "/home/vagrant/program_boot.pdf"
	key := []string{"pro", "boot"}
	p := "<span style=\"background-color:#FFCC00;\">"
	q := "</span>"
	actualf, actuald := r.highlightFilename(s, key)
	expectedf := "<a target=\"_blank\" href=\"file:///home/vagrant/program_boot.pdf\">" + // Link
		"/home/vagrant/" +
		p + "pro" + q +
		"gram_" +
		p + "boot" + q +
		".pdf" +
		"</a>" // Text
	if actualf != expectedf {
		t.Fatalf("got: %v want: %v", actualf, expectedf)
	}
	expectedd := "<a href=\"file:///home/vagrant\" title=\"<< クリックでフォルダに移動\"><<</a>" //Directory
	if actuald != expectedd {
		t.Fatalf("got: %v want: %v", actuald, expectedd)
	}
	/*
		r = Result{}  // has Root Case
		r = Result{}  // Windows Path Case
	*/
}

func Test_highlightString(t *testing.T) {
	s := "/home/vagrant/Program/hoge3/program_boot.pdf"
	actual := highlightString(s, []string{"program", "pdf"})
	p := "<span style=\"background-color:#FFCC00;\">"
	q := "</span>"
	expected := "/home/vagrant/" +
		p + "Program" + q +
		"/hoge3/program_boot." +
		p + "pdf" + q
	if actual != expected {
		t.Fatalf("got: %v want: %v", actual, expected)
	}
}

func TestResult_HTMLContents(t *testing.T) {
	var (
		test = []string{"/home/test/path",
			"this is a test word",
			"0", "1", "2", "3", "4", "5", "6", "7"}
		r        = Result{Out: test}
		key      = []string{"test", "word"}
		p        = "<span style=\"background-color:#FFCC00;\">"
		q        = "</span>"
		actual   = r.HTMLContents(key)
		expected = Content{
			File: fmt.Sprintf("<a target=\"_blank\" href=\"file:///home/test/path\">/home/%stest%s/path</a>", p, q),
			Dir:  "<a href=\"file:///home/test\" title=\"<< クリックでフォルダに移動\"><<</a>",
		}
		expected2 = Content{
			Highlight: fmt.Sprintf("this is a %stest%s %sword%s", p, q, p, q),
		}
		expected3 = []string{"0", "1", "2", "3", "4", "5", "6", "7"}
	)
	if actual.Contents[0].File != expected.File { // Filename test
		t.Fatalf("got: %v want: %v", actual.Contents[0].File, expected.File)
	}
	if actual.Contents[0].Dir != expected.Dir { // Dir test
		t.Fatalf("got: %v want: %v", actual.Contents[0].Dir, expected.Dir)
	}
	if actual.Contents[1].Highlight != expected2.Highlight { // Contents test
		t.Fatalf("got: %v want: %v", actual.Contents[1], expected2.Highlight)
	}
	for i, expect := range expected3 { // Stats test
		if actual.Stats[i] != expect {
			t.Fatalf("got: %v want: %v", actual.Stats[i], expect)
		}
	}
}
