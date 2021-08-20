package cmd

import "testing"

func TestResult_highlightFilename(t *testing.T) {
	r := Result{} // Normal Case
	s := "/home/vagrant/program_boot.pdf"
	actual := r.highlightFilename(s)
	expected := "<a target=\"_blank\" href=\"file:///home/vagrant/program_boot.pdf\"" + // Link
		">/home/vagrant/program_boot.pdf</a>" + // Text
		" <a href=\"file:///home/vagrant\" title=\"<< クリックでフォルダに移動\"><<</a>" //Directory
	if actual != expected {
		t.Fatalf("got: %v want: %v", actual, expected)
	}
	/*
		r = Result{}  // has Root Case
		r = Result{}  // Windows Path Case
	*/
}

func Test_highlightString(t *testing.T) {
	s := "/home/vagrant/Program/hoge3/program_boot.pdf"
	actual := highlightString(s, "program", "pdf")
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
	test := []string{"/home/test/path", "this is a test word", "0", "1", "2", "3", "4", "5", "6", "7"}
	r := Result{Out: test}
	key := "test word"
	actual := r.HTMLContents(key)
	expected := Result{
		Contents: []string{
			highlightString(r.highlightFilename(test[0]), "test", "word"),
			highlightString(test[1], "test", "word"),
		},
	}
	if actual.Contents[0] != expected.Contents[0] { // Filename test
		t.Fatalf("got: %v want: %v", actual.Contents[0], expected.Contents[0])
	}
	if actual.Contents[1] != expected.Contents[1] { // Contents test
		t.Fatalf("got: %v want: %v", actual.Contents[1], expected.Contents[1])
	}
	for i, expect := range []string{"0", "1", "2", "3", "4", "5", "6", "7"} { // Stats test
		if actual.Stats[i] != expect {
			t.Fatalf("got: %v want: %v", actual.Stats[i], expect)
		}
	}
}
