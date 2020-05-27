package main

import "testing"

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
	s := "this is test"
	// and test
	method := "and"
	actual := andorPadding(s, method)
	expected := "this.*is.*test"
	if actual != expected {
		t.Fatalf("got: %v want: %v", actual, expected)
	}
	// OR test
	method = "or"
	actual = andorPadding(s, method)
	expected = "(this|is|test)"
	if actual != expected {
		t.Fatalf("got: %v want: %v", actual, expected)
	}
}
