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
	root         = flag.String("r", "", "DB root directory")
	pathSplitWin = flag.Bool("s", false, "OS path split windows backslash")
)

// PathMap : File:ファイルネームを起点として、
// そのディレクトリと検索語をハイライトした文字列を入れる
type PathMap struct {
	File      string
	Line      string
	Dir       string
	Highlight string
}

func main() {
	// Version info
	flag.BoolVar(&showVersion, "v", false, "show version")
	flag.BoolVar(&showVersion, "version", false, "show version")
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

// htmlClause    : ページに表示する情報
//  searchWord   : 検索キーワード
// directoryPath : ディレクトリパス
//		   depth : Lvを選択したhtml
// 		   andor : and / or 検索方式ラジオボタン
func htmlClause(searchWord, directoryPath, depth, andor string) string {
	return fmt.Sprintf(
		`<!DOCTYPE html>
			<html>
			<head>
			<meta http-equiv="Content-Type" content="text/html; charaset=utf-8">
			<title>Grep Server` + searchWord + directoryPath + `</title>
			</head>
			  <body>
			    <form method="get" action="/search">
				  <!-- directory -->
				  <input type="text"
					  placeholder="検索対象フォルダのフルパスを入力してください(ex:/usr/bin ex:\\gr.net\ShareUsers\User\Personal)"
					  name="directory-path"
					  id="directory-path"
					  value="` + directoryPath + `"
					  size="140"
					  title="検索対象フォルダのフルパスを入力してください(ex:/usr/bin ex:\\gr.net\ShareUsers\User\Personal)">
				  <a href=https://github.com/u1and0/grep-server/blob/master/README.md>Help</a>
				  <br>

				  <!-- file -->
				  <input type="text"
					  placeholder="検索キーワードをスペース区切りで入力してください"
					  name="query"
					  value="` + searchWord + `"
					  size="100"
					  title="検索キーワードをスペース区切りで入力してください">

				   <!-- depth -->
				   Lv
				   <select name="depth"
					  id="depth"
					  size="1"
					  title="Lv: 検索階層数を指定します。数字を増やすと検索速度は落ちますがマッチする可能性が上がります。">
					  ` + depth + `
				  </select>
				 <!-- and/or -->
				 ` + andor + `
				 <input type="submit" name="submit" value="検索" title="スペース区切りをandとみなすかorとみなすか選択します">
			    </form>
				<table>`)
}

// showInit : Top page html
func showInit(w http.ResponseWriter, r *http.Request) {
	// 検索語、ディレクトリは空
	// 検索階層は何もselectされていない(デフォルトは一番上の1になる)
	fmt.Fprintf(w, htmlClause("", "", `
					<option value="1">1</option>
					<option value="2">2</option>
					<option value="3">3</option>
					<option value="4">4</option>
					<option value="5">5</option>
	`,
		`<input type="radio" value="and" name="andor-search" checked="checked">and
		 <input type="radio" value="or"  name="andor-search">or`))
}

// andorPadding : 検索キーワードのスペースをandなら".*" orなら"|"で埋める
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

// addResult : Print ripgrep-all result as html contents
func addResult(w http.ResponseWriter, r *http.Request) {
	// Modify query
	receiveValue := r.FormValue("query")
	directoryPath := r.FormValue("directory-path")
	searchAndOr := r.FormValue("andor-search")
	searchDepth := r.FormValue("depth")
	slashedDirPath := directoryPath
	if *root != "" {
		slashedDirPath = strings.TrimPrefix(slashedDirPath, *root)
	}
	if *pathSplitWin {
		// filepath.ToSlash(directoryPath) <= Windows版Goでしか有効でない
		slashedDirPath = strings.ReplaceAll(slashedDirPath, `\`, "/")
	}

	// コマンド生成
	opt := []string{ // rga/rg options
		"--line-number",
		"--max-columns", "160",
		"--max-columns-preview",
		"--heading",
		"--color", "never",
		"--no-binary",
		"--ignore-case",
		"--max-depth", searchDepth,
	}
	searchWord := andorPadding(receiveValue, searchAndOr)
	opt = append(opt, searchWord)
	opt = append(opt, slashedDirPath)

	// File contents search by `rga` command
	startTime := time.Now()
	out, err := exec.Command("rga", opt...).Output()
	searchTime := float64((time.Since(startTime)).Nanoseconds()) / float64(time.Millisecond)
	if err != nil {
		log.Println(err)
	}

	// 結果をarray型に格納
	outstr := string(out)
	results := strings.Split(outstr, "\n")
	results = results[:len(results)-1] // Pop last element cause \\n

	/* html表示 */
	// 検索後のフォームに再度同じキーワードを入力
	fmt.Fprintf(w, htmlClause(receiveValue, directoryPath,
		// LvDDリスト
		// html上で選択した階層数を記憶して遷移先ページでも同じ数字を選択
		func() string {
			s := `<option value="1">1</option>
				<option value="2">2</option>
				<option value="3">3</option>
				<option value="4">4</option>
				<option value="5">5</option>`
			return strings.Replace(s,
				">"+searchDepth,
				" selected>"+searchDepth,
				1)
		}(),
		// and / or ラジオボタン
		func() string {
			s := `<input type="radio" value="and" name="andor-search">and
				 <input type="radio" value="or"  name="andor-search">or`
			return strings.Replace(s,
				"\"andor-search\">"+searchAndOr,
				"\"andor-search\"checked=\"checked\">"+searchAndOr,
				1) // and かor 選択されている方に"checked"をつける
		}(),
	))
	fmt.Fprintf(w, `<h4> 検索にかかった時間: %.3fmsec </h4>`, searchTime)

	/* 検索結果表示 */
	var contentNum, fileNum int
	match := regexp.MustCompile(`^\d`)
	for _, s := range results {
		if match.MatchString(s) { // 行数から始まるときはfile contents
			fmt.Fprintf(w, // => http.ResponseWriter
				`<tr> <td> %s </td> <tr>`, highlightString(
					html.EscapeString(s),
					// メタ文字含まない検索文字のみhighlight
					strings.Fields(receiveValue)...),
			)
			contentNum++
		} else { // 行数から始まらないときはfile name
			fmt.Fprintf(w, `<tr> <td> %s </td> <tr>`, highlightFilename(s))
			fileNum++
		}
	}
	fmt.Fprintln(w, `</table>
				</body>
				</html>`)

	log.Printf(
		"%4dfiles %6dmatched lines %3.3fmsec Keyword: [ %-30s ] Path: [ %-50s ]\n",
		fileNum, contentNum, searchTime, searchWord, directoryPath)
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
		re := regexp.MustCompile(`((?i)` + w + `)`)
		found := re.FindString(s)
		if found != "" {
			s = strings.Replace(s, found,
				"<span style=\"background-color:#FFCC00;\">"+found+"</span>", -1)
		}
	}
	return s
}
