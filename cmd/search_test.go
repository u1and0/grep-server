package cmd

import (
	"errors"
	"testing"
	"time"
)

func TestSearch_Grep(t *testing.T) {
	var err error
	// No keyrord
	s := Search{}
	expectedErr := errors.New("検索キーワードがありません")
	if _, err = s.Grep(); err == nil {
		t.Errorf("This test must fail.\ngot: %v want: %v", err, expectedErr)
	}
	// No path
	s = Search{Keyword: "hoge"}
	expectedErr = errors.New("ディレクトリパスがありません")
	if _, err = s.Grep(); err == nil {
		t.Errorf("This test must fail.\ngot: %v want: %v", err, expectedErr)
	}
	// Not exist path
	s = Search{Keyword: "hoge", Path: "fuga", CmdPath: "foo"}
	expectedErr = errors.New("ディレクトリパス " + s.CmdPath + " がありません")
	if _, err = s.Grep(); err == nil {
		t.Errorf("This test must fail.\ngot: %v want: %v", err, expectedErr)
	}
	// Timeout
	s = Search{
		Keyword:    "hoge",
		Path:       "/var/log",
		AndOr:      "and",
		Depth:      "5",
		Encoding:   "UTF-8",
		CmdKeyword: "hoge",
		CmdPath:    "foo",
		Exe:        "/usr/bin/rga",
		Timeout:    1 * time.Millisecond,
	}
	expectedErr = errors.New("タイムアウトしました。検索条件を変えてください。")
	if _, err = s.Grep(); err == nil {
		t.Errorf("This test must fail.\ngot: %v want: %v", err, expectedErr)
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
