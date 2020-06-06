package main

import (
	"errors"
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
)

const (
	// VERSION : version
	VERSION = "0.0.0"
	// EXE : Search command
	EXE = "/usr/bin/rga"
	// LOGFILE : 検索条件 / マッチファイル数 / マッチ行数 / 検索時間を記録するファイル
	LOGFILE = "/var/log/grep-server.log"
	// STATSLENGTH : rg --stats の行数
	STATSLENGTH = 8
)

var (
	showVersion  bool
	debug        bool
	root         = flag.String("r", "", "Append root directory path")
	encoding     = flag.String("E", "UTF-8", "Set default encoding")
	pathSplitWin = flag.Bool("s", false, "OS path split windows backslash")
	result       = Result{}
)

// Search : Search query structure
type Search struct {
	Keyword    string //  検索語
	Path       string //  検索対象パス
	AndOr      string //  and / or の検索メソッド
	Depth      string //  検索対象パスから検索する階層数
	Encoding   string //  ファイルエンコード
	CmdKeyword string //  rgaコマンドに渡す and / or padding した検索キーワード
	CmdPath    string //  rgaコマンドに渡す'/'に正規化し、ルートパスを省いたパス
}

// Result : rga結果, Statsと結果に別れる
type Result struct {
	Stats    []string
	Contents []string
}

func main() {
	// Version info
	flag.BoolVar(&showVersion, "v", false, "show version")
	flag.BoolVar(&showVersion, "version", false, "show version")
	flag.BoolVar(&debug, "debug", false, "run as debug mode")
	flag.Parse()
	if showVersion {
		fmt.Println("grep-server", VERSION)
		rgaVersion, _ := exec.Command(EXE, "--version").Output()
		fmt.Println(string(rgaVersion))
		return // versionを表示して終了
	}
	// Command check
	if _, err := exec.LookPath(EXE); err != nil {
		log.Fatalf("[ERROR] %s", err.Error())
	}
	// Log setting
	logfile, err := os.OpenFile(LOGFILE, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("[ERROR] %s", err.Error())
	}
	defer logfile.Close()
	log.SetOutput(io.MultiWriter(logfile, os.Stdout))
	// HTTP response
	http.HandleFunc("/", showInit)        // top page
	http.HandleFunc("/search", addResult) // search result
	http.ListenAndServe(":8080", nil)
}

// htmlClause : ページに表示する情報
func (s *Search) htmlClause() string {
	pathtext := `"検索対象フォルダのフルパスを入力してください(ex:/usr/bin ex:\\gr.net\ShareUsers\User\Personal)"`
	keytext := `"検索キーワードをスペース区切りで入力してください"`
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
					title="文字エンコードを指定します。">
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

// showInit : Top page html
func showInit(w http.ResponseWriter, r *http.Request) {
	// 検索語、ディレクトリは空
	// 検索階層は何もselectされていない(デフォルトは一番上の1になる)
	s := Search{Depth: "1", AndOr: "and", Encoding: *encoding}
	if debug {
		fmt.Printf("[DEBUG] search struct: %+v\n", s)
	}
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
		log.Fatalf("[ERROR] an error format selected %s."+
			" Must be and/or either.", method)
	}
	return s
}

// システムからbyteで返される結果をsrting リストに格納する
func splitOutByte(b []byte) (a []string) {
	a = strings.Split(string(b), "\n")
	a = a[:len(a)-1] // Pop last element cause \\n
	return
}

func (s *Search) grep() (outstr []string, err error) {
	if *root != "" {
		s.CmdPath = strings.TrimPrefix(s.CmdPath, *root)
	}
	if *pathSplitWin {
		// filepath.ToSlash(Path) <= Windows版Goでしか有効でない
		s.CmdPath = strings.ReplaceAll(s.CmdPath, `\`, "/")
	}
	if s.Keyword == "" {
		return outstr, errors.New("検索キーワードがありません")
	}
	if s.Path == "" { // Directory check
		return outstr, errors.New("ディレクトリパスがありません")
	}
	if _, err = os.Stat(s.CmdPath); os.IsNotExist(err) {
		return outstr, errors.New("ディレクトリパス " + s.CmdPath + " がありません")
	}
	s.CmdKeyword = andorPadding(s.Keyword, s.AndOr)
	if debug {
		fmt.Printf("[DEBUG] search struct: %+v\n", s)
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
		"--stats",
		"--max-depth", s.Depth,
		"--encoding", s.Encoding,

		s.CmdKeyword,
		s.CmdPath,
	}
	if debug {
		fmt.Printf("[DEBUG] options: %v\n", opt)
	}

	// File contents search by `rga` command
	var out []byte
	out, err = exec.Command(EXE, opt...).Output()
	if err != nil {
		log.Printf("[ERROR] %s", err)
	}
	outstr = splitOutByte(out)
	if debug {
		fmt.Printf("[DEBUG] result: %+v\n", outstr)
	}
	return outstr, err
}

// addResult : Print ripgrep-all result as html contents
func addResult(w http.ResponseWriter, r *http.Request) {
	// Modify query
	search := Search{
		Keyword:    r.FormValue("query"),
		Path:       r.FormValue("directory-path"),
		AndOr:      r.FormValue("andor-search"),
		Depth:      r.FormValue("depth"),
		Encoding:   r.FormValue("encoding"),
		CmdKeyword: "",
		CmdPath:    r.FormValue("directory-path"), // 初期値はPathと同じ
	}
	if debug {
		fmt.Printf("[DEBUG] search struct: %+v\n", search)
	}

	/* html表示 */
	fmt.Fprintf(w, search.htmlClause())     // 検索後のフォームに再度同じキーワードを入力
	defer fmt.Fprintf(w, `</body> </html>`) // 終了タグ
	/* 検索結果表示 */
	outstr, err := search.grep()
	if err != nil {
		fmt.Fprintf(w, `<h4> %s </h4>`, err)
		log.Printf(
			"[ERROR] %s Keyword: [ %-30s ] Path: [ %-50s ]\n",
			err, search.Keyword, search.Path)
	} else {
		result = htmlContents(outstr, search.Keyword)
		fmt.Fprintf(w, "<h4>")
		for _, h := range result.Stats {
			fmt.Fprintf(w, h)
			fmt.Fprintf(w, "<br>")
		}
		fmt.Fprintf(w, "</h4>")
		fmt.Fprintf(w, `<table>`)
		for _, h := range result.Contents {
			fmt.Fprintf(w, `<tr> <td>`+h+`</td> </tr>`)
		}
		fmt.Fprintf(w, `</table>`)
		log.Printf(
			"%s Keyword: [ %-30s ] Path: [ %-50s ]\n",
			strings.Join(result.Stats, " "), search.Keyword, search.Path)
	}
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

func htmlContents(a []string, key string) (r Result) {
	var (
		l = len(a) - STATSLENGTH
		x = regexp.MustCompile(`^/`)
		h string // highlight string
	)
	for _, s := range a[:l] {
		if x.MatchString(s) { // '/'から始まるときはfilename
			h = highlightFilename(s)
		} else { // '/'から始まらないときはfile contents
			h = highlightString(
				html.EscapeString(s),
				// メタ文字含まない検索文字のみhighlight
				strings.Fields(key)...)
		}
		r.Contents = append(r.Contents, h)
	}
	r.Stats = a[l:]
	return
}
