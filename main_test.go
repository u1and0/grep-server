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

func Test_htmlClause(t *testing.T) {
	// Case 1
	s := Search{Depth: "1", AndOr: "and"}
	actual := s.htmlClause()
	expected := `<!DOCTYPE html>
			<html>
			<head>
			<meta http-equiv="Content-Type" content="text/html; charaset=utf-8">
			<title>Grep Server  </title>
			</head>
			  <body>
			    <form method="get" action="/search">
				  <!-- directory -->
				  <input type="text"
					  placeholder="検索対象フォルダのフルパスを入力してください(ex:/usr/bin ex:\\gr.net\ShareUsers\User\Personal)"
					  name="directory-path"
					  id="directory-path"
					  value=""
					  size="140"
					  title="検索対象フォルダのフルパスを入力してください(ex:/usr/bin ex:\\gr.net\ShareUsers\User\Personal)">
				  <a href=https://github.com/u1and0/grep-server/blob/master/README.md>Help</a>
				  <br>

				  <!-- file -->
				  <input type="text"
					  placeholder="検索キーワードをスペース区切りで入力してください"
					  name="query"
					  value=""
					  size="100"
					  title="検索キーワードをスペース区切りで入力してください">

				   <!-- depth -->
				   Lv
				   <select name="depth"
					  id="depth"
					  size="1"
					  title="Lv: 検索階層数を指定します。数字を増やすと検索速度は落ちますがマッチする可能性が上がります。">
					<option value="1" selected>1</option>
					<option value="2">2</option>
					<option value="3">3</option>
					<option value="4">4</option>
					<option value="5">5</option>
				  </select>
				 <!-- and/or -->
				 <input type="radio" value="and"
					title="スペース区切りをandとみなすかorとみなすか選択します"
					name="andor-search" checked="checked">and
					<input type="radio" value="or"
					title="スペース区切りをandとみなすかorとみなすか選択します"
					name="andor-search">or
				 <input type="submit" name="submit" value="検索">
			    </form>
				<table>`
	if actual != expected {
		t.Fatalf("got: %v want: %v", actual, expected)
	}

	// Case 2
	s = Search{
		Keyword: "test word",
		Path:    "/home/testuser",
		Depth:   "3",
		AndOr:   "and",
	}
	actual = s.htmlClause()
	expected = `<!DOCTYPE html>
			<html>
			<head>
			<meta http-equiv="Content-Type" content="text/html; charaset=utf-8">
			<title>Grep Server test word /home/testuser</title>
			</head>
			  <body>
			    <form method="get" action="/search">
				  <!-- directory -->
				  <input type="text"
					  placeholder="検索対象フォルダのフルパスを入力してください(ex:/usr/bin ex:\\gr.net\ShareUsers\User\Personal)"
					  name="directory-path"
					  id="directory-path"
					  value="/home/testuser"
					  size="140"
					  title="検索対象フォルダのフルパスを入力してください(ex:/usr/bin ex:\\gr.net\ShareUsers\User\Personal)">
				  <a href=https://github.com/u1and0/grep-server/blob/master/README.md>Help</a>
				  <br>

				  <!-- file -->
				  <input type="text"
					  placeholder="検索キーワードをスペース区切りで入力してください"
					  name="query"
					  value="test word"
					  size="100"
					  title="検索キーワードをスペース区切りで入力してください">

				   <!-- depth -->
				   Lv
				   <select name="depth"
					  id="depth"
					  size="1"
					  title="Lv: 検索階層数を指定します。数字を増やすと検索速度は落ちますがマッチする可能性が上がります。">
					<option value="1">1</option>
					<option value="2">2</option>
					<option value="3" selected>3</option>
					<option value="4">4</option>
					<option value="5">5</option>
				  </select>
				 <!-- and/or -->
				 <input type="radio" value="and"
					title="スペース区切りをandとみなすかorとみなすか選択します"
					name="andor-search" checked="checked">and
					<input type="radio" value="or"
					title="スペース区切りをandとみなすかorとみなすか選択します"
					name="andor-search">or
				 <input type="submit" name="submit" value="検索">
			    </form>
				<table>`
	if actual != expected {
		t.Fatalf("got: %v want: %v", actual, expected)
	}
}
