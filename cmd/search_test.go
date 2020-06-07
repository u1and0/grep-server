package cmd

import "testing"

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
