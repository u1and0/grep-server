package cmd

import (
	"errors"
	"testing"
	"time"
)

func TestSearch_Grep(t *testing.T) {
	var err error
	// No keyrord Error
	s := Search{}
	expectedErr := errors.New("検索キーワードがありません")
	if _, err = s.Grep(); err == nil {
		t.Errorf("This test must fail.\ngot: %v want: %v", err, expectedErr)
	}
	// No path Error
	s = Search{Keyword: "hoge"}
	expectedErr = errors.New("ディレクトリパスがありません")
	if _, err = s.Grep(); err == nil {
		t.Errorf("This test must fail.\ngot: %v want: %v", err, expectedErr)
	}
	// Not exist path Error
	s = Search{Keyword: "hoge", Path: "fuga", CmdPath: "foo"}
	expectedErr = errors.New("ディレクトリパス " + s.CmdPath + " がありません")
	if _, err = s.Grep(); err == nil {
		t.Errorf("This test must fail.\ngot: %v want: %v", err, expectedErr)
	}
	// Timeout Error
	s = Search{
		Keyword:  "bash",
		Path:     "../test",
		AndOr:    "and",
		Depth:    "1",
		Encoding: "UTF-8",
		CmdPath:  "../test",
		Exe:      "/usr/bin/rga",
		Timeout:  1 * time.Nanosecond,
	}
	expectedErr = errors.New("タイムアウトしました。検索条件を変えてください。")
	if _, err = s.Grep(); err == nil {
		t.Errorf("This test must fail.\ngot: %v want: %v", err, expectedErr)
	}
	// Success
	s = Search{
		Keyword:  "bash",
		Path:     "../test",
		AndOr:    "and",
		Depth:    "1",
		Encoding: "UTF-8",
		CmdPath:  "../test",
		Exe:      "/usr/bin/rga",
		Timeout:  10 * time.Second,
	}
	expected := 227 + // Matched lines
		8 + // Stats lines
		2 + // Filenames
		2 // CRLF
	actual, _ := s.Grep()
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
