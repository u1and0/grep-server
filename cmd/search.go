package cmd

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Search : Search query structure
type Search struct {
	Keyword      string        // 検索語
	Path         string        // 検索対象パス
	AndOr        string        // and / or の検索メソッド
	Depth        string        // 検索対象パスから検索する階層数
	Encoding     string        // ファイルエンコード
	Mode         string        // 検索モード Content ファイル内容 / File ファイル名
	CmdKeyword   string        // rgaコマンドに渡す and / or padding した検索キーワード
	CmdPath      string        // rgaコマンドに渡す'/'に正規化し、ルートパスを省いたパス
	Exe          string        // => /usr/bin/rga
	Exf          string        // => /usr/bin/rg
	Root         string        // 追加するパスのプレフィックス
	Trim         string        // 取り除くパスのプレフィックス
	PathSplitWin bool          // Windows path sepに変更する
	Debug        bool          // Debug modeに変更する
	Timeout      time.Duration // rga検索コマンドのタイムアウト
}

// CommandGen : rgaオプション設定
func (s *Search) CommandGen() ([]string, error) {
	var command []string
	if s.Root != "" {
		s.CmdPath = strings.TrimPrefix(s.CmdPath, s.Root)
	}
	if s.Trim != "" {
		s.CmdPath = s.Trim + s.CmdPath
	}
	if s.PathSplitWin {
		// filepath.ToSlash(Path) <= Windows版Goでしか有効でない
		s.CmdPath = strings.ReplaceAll(s.CmdPath, `\`, "/")
	}
	if s.Keyword == "" {
		return []string{}, errors.New("検索キーワードがありません")
	}
	if s.Path == "" { // Directory check
		return []string{}, errors.New("ディレクトリパスがありません")
	}
	if _, err := os.Stat(s.CmdPath); os.IsNotExist(err) {
		return []string{}, errors.New("ディレクトリパス " + s.CmdPath + " がありません")
	}
	s.CmdKeyword = andorPadding(s.Keyword, s.AndOr)
	if s.Debug {
		fmt.Printf("[DEBUG] search struct: %+v\n", s)
	}

	// コマンド生成
	if s.Mode == "Content" {
		command = []string{
			"--line-number",
			"--max-columns", "160",
			"--max-columns-preview",
			"--heading",
			"--color", "never",
			"--no-binary",
			"--smart-case",
			// "--ignore-case",
			"--stats",
			"--max-depth", s.Depth,
			"--encoding", s.Encoding,

			s.CmdKeyword,
			s.CmdPath,
		}
	} else if s.Mode == "File" {
		command = []string{
			"--files",
			"--max-depth", s.Depth,
			s.CmdPath,
			"|",
			s.Exf,
			"--color", "never",
			"--smart-case",
			// "--ignore-case",
			"--stats",
			fmt.Sprintf("\"%s\"", s.CmdKeyword),
			// or 検索で "|" が入るとパイプとみなされるため
			// Double quote for Escaping
		}
	} else {
		return []string{}, errors.New("検索モードが設定されていません。 Content or File")
	}
	return command, nil
}

// Grep : rga検索の結果をstring sliceにして返す
func (s *Search) Grep(opt []string) ([]string, error) {
	// File contents search by `rga` command
	var (
		out []byte
		err error
	)
	if s.Mode == "Content" {
		// Create a new context and add a timeout to it
		ctx, cancel := context.WithTimeout(context.Background(), s.Timeout)
		defer cancel() // The cancel should be deferred so resources are cleaned up
		command := exec.CommandContext(ctx, s.Exe, opt...)
		out, err = command.Output()
		// We want to check the context error to see if the timeout was executed.
		// The error returned by cmd.Output() will be OS specific based on what
		// happens when a process is killed.
		if ctx.Err() == context.DeadlineExceeded {
			return []string{}, errors.New("タイムアウトしました。検索条件を変えてください。")
		}
	} else if s.Mode == "File" {
		command := exec.Command("sh", "-c", s.Exf+" "+strings.Join(opt, " "))
		out, err = command.CombinedOutput()
	} else {
		log.Fatalf("[ERROR] an error format selected %s."+
			" Must be Content/File either.", s.Mode)
	}
	outstr := splitOutByte(out)
	return outstr, err
}

// HTMLClause : ページに表示する情報
func (s *Search) HTMLClause() string {
	pathtext := ``
	keytext := ``
	return fmt.Sprintf(
		`<!DOCTYPE html>
			<html>
			<head>
			<meta http-equiv="Content-Type" content="text/html; charaset=utf-8">
			<title>` + strings.Join([]string{"Grep Server", s.Keyword, s.Path}, " ") + `</title>
			</head>
			  <body>
			    <form method="get" action="/search">

				  <!-- directory -->
				  <input type="text"
					  placeholder=` + pathtext + `
					  name="directory-path"
					  id="directory-path"
					  value="` + s.Path + `"
					  size="140"
					  title=` + pathtext + `>
				  <a href=https://github.com/u1and0/grep-server/blob/master/README.md>Help</a>
				  <br>

				  <!-- file -->
				  <input type="text"
					  placeholder=` + keytext + `
					  name="query"
					  value="` + s.Keyword + `"
					  size="90"
					  title=` + keytext + `>

				   <!-- mode -->
				   <select name="mode"
					id="mode"
					size="1"
					title="Mode: 検索モードを指定します。ファイル内検索したいとき=>'Content'を選択, ファイル名検索をしたいとき=>'File'を選択してください。">
					` +
			func() string { // 検索モードは何もselectされていない(デフォルトは一番上のContentになる)
				n := `<option value="Content">Content</option>
					<option value="File">File</option>`
				return strings.Replace(n, ">"+s.Mode, " selected>"+s.Mode, 1)
			}() + `
				  </select>

				   <!-- depth -->
				   Lv
				   <select name="depth"
					  id="depth"
					  size="1"
					  title="Lv: 検索階層数を指定します。数字を増やすと検索速度は落ちますがマッチする可能性が上がります。">
					` +
			func() string { // 検索階層は何もselectされていない(デフォルトは一番上の1になる)
				n := `<option value="1">1</option>
					<option value="2">2</option>
					<option value="3">3</option>
					<option value="4">4</option>
					<option value="5">5</option>`
				return strings.Replace(n, ">"+s.Depth, " selected>"+s.Depth, 1)
			}() + `
				  </select>

				 <!-- and/or -->
				 ` +
			func() string { // and かor 選択されている方に"checked"をつける
				n := `<input type="radio" value="and"
					title="スペース区切りをandとみなすかorとみなすか選択します"
					name="andor-search">and
					<input type="radio" value="or"
					title="スペース区切りをandとみなすかorとみなすか選択します"
					name="andor-search">or`
				return strings.Replace(n,
					"\"andor-search\">"+s.AndOr,
					"\"andor-search\" checked=\"checked\">"+s.AndOr,
					1)
			}() + `

				 <!-- encoding -->
				 <select name="encoding"
					id="encoding"
					size="1"
					title="文字エンコードを指定します。結果が文字化けするときはリストから適宜選択してください。">
				` +
			func() string { // 文字エンコーディングはデフォルトUTF-8
				n := `<option value="UTF-8">UTF-8</option>
					<option value="SHIFT-JIS">SHIFT-JIS</option>
					<option value="EUC-JP">EUC-JP</option>
					<option value="ISO-2022-JP">ISO-2022-JP</option>`
				return strings.Replace(n, ">"+s.Encoding, " selected>"+s.Encoding, 1)
			}() + `
				  </select>
				 ` +
			`<input type="submit" name="submit" value="Search">
			    </form>`)
}

// andorPadding : 検索キーワードをrgaコマンドへ渡す形式に正規化する
func andorPadding(s, method string) string {
	ss := strings.Fields(s)
	if method == "and" {
		method = ".*"
		s = strings.Join(ss, method)
	} else if method == "or" {
		method = "|"
		s = strings.Join(ss, method)
		s = "(" + s + ")"
	} else {
		log.Fatalf("[ERROR] an error format selected %s."+
			" Must be and/or either.", method)
	}
	return s
}

// splitOutByte : システムからbyteで返される結果をsrting リストに格納する
func splitOutByte(b []byte) (a []string) {
	a = strings.Split(string(b), "\n")
	a = a[:len(a)-1] // Pop last element cause \\n
	return
}
