package main

import (
	"flag"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	// VERSION : version
	VERSION = "0.0.0"
	// LOGFILE : 検索条件 / マッチファイル数 / マッチ行数 / 検索時間を記録するファイル
	LOGFILE = "/var/log/grep-server.log"
)

var (
	showVersion  bool
	debug        bool
	root         = flag.String("r", "", "DB root directory")
	pathSplitWin = flag.Bool("s", false, "OS path split windows backslash")
)

// Search : Search query structure
type Search struct {
	Keyword    string //  検索語
	Path       string //  検索対象パス
	AndOr      string //  and / or の検索メソッド
	Depth      string //  検索対象パスから検索する階層数
	CmdKeyword string //  rgaコマンドに渡す and / or padding した検索キーワード
	CmdPath    string //  rgaコマンドに渡す'/'に正規化し、ルートパスを省いたパス
}

// Match : Matched contents
type Match struct{ Line, File int }

/*
// PathMap : File:ファイルネームを起点として、
// そのディレクトリと検索語をハイライトした文字列を入れる
type PathMap struct {
	File      string
	Line      string
	Dir       string
	Highlight string
}
*/

func main() {
	// Version info
	flag.BoolVar(&showVersion, "v", false, "show version")
	flag.BoolVar(&showVersion, "version", false, "show version")
	flag.BoolVar(&debug, "debug", false, "run as debug mode")
	flag.Parse()
	if showVersion {
		fmt.Println("grep-server", VERSION)
		return // versionを表示して終了
	}
	// Command check
	if _, err := exec.LookPath("rga"); err != nil {
		log.Fatal(err)
	}
	// Log setting
	logfile, err := os.OpenFile(LOGFILE, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("[ERROR] Cannot open logfile " + err.Error())
	}
	defer logfile.Close()
	log.SetOutput(io.MultiWriter(logfile, os.Stdout))
	// HTTP response
	http.HandleFunc("/", showInit)        // top page
	http.HandleFunc("/search", addResult) // search result
	http.ListenAndServe(":8080", nil)
}

// htmlClause : ページに表示する情報
// depth	  : Lvを選択したhtml
// andor 	  : and / or 検索方式ラジオボタン
func (s *Search) htmlClause() string {
	pathtext := `検索対象フォルダのフルパスを入力してください(ex:/usr/bin ex:\\gr.net\ShareUsers\User\Personal)`
	keytext := `検索キーワードをスペース区切りで入力してください`
	return fmt.Sprintf(
		`<!DOCTYPE html>
			<html>
			<head>
			<meta http-equiv="Content-Type" content="text/html; charaset=utf-8">
			<title>Grep Server` + s.Keyword + s.Path + `</title>
			</head>
			  <body>
			    <form method="get" action="/search">
				  <!-- directory -->
				  <input type="text"
					  placeholder="` + pathtext + `"
					  name="directory-path"
					  id="directory-path"
					  value="` + s.Path + `"
					  size="140"
					  title="` + pathtext + `">
				  <a href=https://github.com/u1and0/grep-server/blob/master/README.md>Help</a>
				  <br>

				  <!-- file -->
				  <input type="text"
					  placeholder=` + keytext + `
					  name="query"
					  value="` + s.Keyword + `"
					  size="100"
					  title="` + keytext + `">

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
				n := `
				<input type="radio" value="and" name="andor-search"
				title="スペース区切りをandとみなすかorとみなすか選択します">and
				<input type="radio" value="or"  name="andor-search"
				title="スペース区切りをandとみなすかorとみなすか選択します">or
				`
				return strings.Replace(n,
					"\"andor-search\">"+s.AndOr,
					"\"andor-search\"checked=\"checked\">"+s.AndOr,
					1)
			}() + `
				 <input type="submit" name="submit" value="検索">
			    </form>
				<table>`)
}

// showInit : Top page html
func showInit(w http.ResponseWriter, r *http.Request) {
	// 検索語、ディレクトリは空
	// 検索階層は何もselectされていない(デフォルトは一番上の1になる)
	s := Search{Depth: "1", AndOr: "and"}
	fmt.Fprintf(w, s.htmlClause())
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
		log.Fatalf("an error format selected %s", method)
	}
	return s
}

// システムからbyteで返される結果をsrting リストに格納する
func splitOutByte(b []byte) []string {
	results := strings.Split(string(b), "\n")
	results = results[:len(results)-1] // Pop last element cause \\n
	return results
}

// addResult : Print ripgrep-all result as html contents
func addResult(w http.ResponseWriter, r *http.Request) {
	// Modify query
	search := Search{
		Keyword:    r.FormValue("query"),
		Path:       r.FormValue("directory-path"),
		AndOr:      r.FormValue("andor-search"),
		Depth:      r.FormValue("depth"),
		CmdKeyword: "",
		CmdPath:    r.FormValue("directory-path"), // 初期値はPathと同じ
	}
	if *root != "" {
		search.CmdPath = strings.TrimPrefix(search.Path, *root)
	}
	if *pathSplitWin {
		// filepath.ToSlash(Path) <= Windows版Goでしか有効でない
		search.CmdPath = strings.ReplaceAll(search.Path, `\`, "/")
	}
	search.CmdKeyword = andorPadding(search.Keyword, search.AndOr)
	if debug {
		fmt.Printf("[DEBUG] search struct: %v\n", search)
	}

	// コマンド生成
	opt := []string{ // rga/rg options
		"--line-number",
		"--max-columns", "160",
		"--max-columns-preview",
		"--heading",
		"--color", "never",
		"--no-binary",
		"--smart-case",
		// "--ignore-case",
		"--max-depth", search.Depth,
		"--stats",

		search.CmdKeyword,
		search.CmdPath,
	}
	if debug {
		fmt.Printf("[DEBUG] options: %v\n", opt)
	}

	// File contents search by `rga` command
	startTime := time.Now()
	out, err := exec.Command("rga", opt...).Output()
	searchTime := float64((time.Since(startTime)).Nanoseconds()) / float64(time.Millisecond)
	if err != nil {
		log.Println(err)
	}
	results := splitOutByte(out)

	/* html表示 */
	// 検索後のフォームに再度同じキーワードを入力
	fmt.Fprintf(w, search.htmlClause())
	fmt.Fprintf(w, `<h4> 検索にかかった時間: %.3fmsec </h4>`, searchTime)

	if debug {
		fmt.Printf("[DEBUG] result: %v\n", results)
	}
	/* 検索結果表示 */
	match := Match{}
	// var contentNum, fileNum int
	regex := regexp.MustCompile(`^[/\\]`)
	for _, s := range results {
		if regex.MatchString(s) { // '/'から始まるときはfilename
			fmt.Fprintf(w, `<tr> <td> %s </td> <tr>`, highlightFilename(s))
			match.File++
			match.Line-- // --heading によりファイル名の前に改行が入るため
		} else { // '/'から始まらないときはfile contents
			fmt.Fprintf(w, // => http.ResponseWriter
				`<tr> <td> %s </td> <tr>`, highlightString(
					html.EscapeString(s),
					// メタ文字含まない検索文字のみhighlight
					strings.Fields(search.Keyword)...),
			)
			match.Line++
		}
	}
	match.Line -= 8 // --stats optionによる行数をマイナスカウント
	fmt.Fprintln(w, `</table>
				</body>
				</html>`)

	log.Printf(
		"%4dfiles %6dmatched lines %3.3fmsec Keyword: [ %-30s ] Path: [ %-50s ]\n",
		match.File, match.Line, searchTime, search.Keyword, search.Path)
}

// ファイル名をリンク化したhtmlを返す
func highlightFilename(s string) string {
	dirpath := filepath.Dir(s)

	// Add drive path
	if *root != "" && s != "" {
		s = *root + s
		dirpath = *root + dirpath
	}
	// windows path convert
	if *pathSplitWin {
		s = strings.ReplaceAll(s, "/", `\`)
	}

	if s != "" {
		s = strings.Replace(s, s,
			"<a target=\"_blank\" href=\"file://"+s+"\">"+s+"</a>", 1)
		s += " <a href=\"file://" + dirpath + "\" title=\"<< クリックでフォルダに移動\"><<</a>"
	}
	return s
}

// highlightString : sの文字列中にあるwordsの背景を黄色にハイライトしたhtmlを返す
func highlightString(s string, words ...string) string {
	for _, w := range words {
		re := regexp.MustCompile(`((?i)` + w + `)`) // ((?i)word)
		found := re.FindString(s)
		if found != "" {
			s = strings.ReplaceAll(s, found,
				"<span style=\"background-color:#FFCC00;\">"+found+"</span>")
		}
	}
	return s
}
