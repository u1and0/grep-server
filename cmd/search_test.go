package cmd

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestSearch_CommandGen(t *testing.T) {
	var err error
	// No keyrord Error
	s := Search{}
	expectedErr := errors.New("検索キーワードがありません")
	if _, err = s.CommandGen(); err == nil {
		t.Errorf("This test must fail.\ngot: %v want: %v", err, expectedErr)
	}
	// No path Error
	s = Search{Keyword: "hoge"}
	expectedErr = errors.New("ディレクトリパスがありません")
	if _, err = s.CommandGen(); err == nil {
		t.Errorf("This test must fail.\ngot: %v want: %v", err, expectedErr)
	}
	// Not exist path Error
	s = Search{Keyword: "hoge", Path: "fuga", CmdPath: "foo"}
	expectedErr = errors.New("ディレクトリパス " + s.CmdPath + " がありません")
	if _, err = s.CommandGen(); err == nil {
		t.Errorf("This test must fail.\ngot: %v want: %v", err, expectedErr)
	}
	// Success
	s = Search{
		Keyword:  "bash",
		Path:     "../test",
		AndOr:    "and",
		Depth:    "1",
		Encoding: "UTF-8",
		Mode:     "Content",
		CmdPath:  "../test",
		Exe:      "/usr/bin/rga",
		Timeout:  10 * time.Second,
	}
	e := []string{
		"--line-number",
		"--max-columns", "160",
		"--max-columns-preview",
		"--heading",
		"--color", "never",
		"--no-binary",
		"--smart-case",
		"--stats",
		"--max-depth", "1",
		"--encoding", "UTF-8",
		"bash",
		"../test",
	}
	a, _ := s.CommandGen()
	actual := strings.Join(a, " ")
	expected := strings.Join(e, " ")
	if expected != actual {
		t.Fatalf("\ngot: %v\nwant: %v", actual, expected)
	}
}

func Test_Grep(t *testing.T) {
	var err error
	// Timeout Error
	s := Search{
		Keyword:  "bash",
		Path:     "../test",
		AndOr:    "and",
		Depth:    "1",
		Encoding: "UTF-8",
		Mode:     "Content",
		CmdPath:  "../test",
		Exe:      "/usr/bin/rga",
		Timeout:  1 * time.Nanosecond,
	}
	c := []string{
		"--line-number",
		"--max-columns", "160",
		"--max-columns-preview",
		"--heading",
		"--color", "never",
		"--no-binary",
		"--smart-case",
		"--stats",
		"--max-depth", "1",
		"--encoding", "UTF-8",

		"bash",    // s.CmdKeyword,
		"../test", // s.CmdPath,
	}
	expectedErr := errors.New("タイムアウトしました。検索条件を変えてください。")
	if _, err = s.Grep(c); err == nil {
		t.Errorf("This test must fail.\ngot: %v want: %v", err, expectedErr)
	}
	// Success
	s = Search{
		// Keyword:  "bash",
		// Path:     "../test",
		AndOr:    "and",
		Depth:    "1",
		Encoding: "UTF-8",
		Mode:     "Content",
		CmdPath:  "../test",
		Exe:      "/usr/bin/rga",
		Timeout:  10 * time.Second,
	}

	expected := 449 + // Matched lines
		8 + // Stats lines
		3 + // Filenames
		3 // CRLF
	actual, _ := s.Grep(c)
	if expected != len(actual) {
		t.Fatalf("got: %v want: %v", len(actual), expected)
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
