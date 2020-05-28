package main

import "testing"

func Test_highlightFilename(t *testing.T) {
	s := "/home/vagrant/program_boot.pdf"
	actual := highlightFilename(s)
	expected := "<a target=\"_blank\" href=\"file:///home/vagrant/program_boot.pdf\"" + // Link
		">/home/vagrant/program_boot.pdf</a>" + // Text
		" <a href=\"file:///home/vagrant\" title=\"<< クリックでフォルダに移動\"><<</a>" //Directory
	if actual != expected {
		t.Fatalf("got: %v want: %v", actual, expected)
	}
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

func Test_andorPadding(t *testing.T) {
	for i, method := range []string{"and", "or"} {
		actual := andorPadding("this is test", method)
		expected := []string{
			"this.*is.*test", // AND Result
			"(this|is|test)", // OR Result
		}
		if actual != expected[i] {
			t.Fatalf("got: %v want: %v", actual, expected)
		}
	}
}

func Test_splitOutByte(t *testing.T) {
	b := []byte("hello\nmy\nname\n") // need last CRLF
	actual := splitOutByte(b)
	expected := []string{"hello", "my", "name"}
	for i, s := range actual {
		if s != expected[i] {
			t.Fatalf("got: %v want: %v", actual, expected)
		}
	}
}
