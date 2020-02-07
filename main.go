package main

import (
	"flag"
	"fmt"
	"html"
	"net/http"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	// VERSION : version
	VERSION = "0.0.0"
	// LOGFILE : 検索条件 / マッチファイル数 / マッチ行数 / 検索時間を記録するファイル
	LOGFILE = "/var/log/gerp-server.log"
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

// htmlClause  : ページに表示する情報
//			 s : 検索キーワード
// 			 d : ディレクトリパス
// 			de : 検索階層数を選択したhtml
func htmlClause(s, d, de, ao string) string {
	return fmt.Sprintf(
		`<!DOCTYPE html>
			<html>
			<head>
			<meta http-equiv="Content-Type" content="text/html; charaset=utf-8">
			<title>Grep Server %s %s</title>
			</head>
			  <body>
			    <form method="get" action="/searching">
				  <input type="text"
					  placeholder="フォルダパス(ex:/usr/bin ex:\ShareUsers\User\Personal)"
					  name="directory-path"
					  id="directory-path"
					  value="%s"
					  size="140"
					  title="フォルダパス">
				  <a href=https://github.com/u1and0/grep-server/blob/master/README.md>Help</a>
				  <br>
				  <input type="text"
					  placeholder="検索語"
					  name="query"
					  value="%s"
					  size="100"
					  title="検索ワード">
				  %s
				  検索階層数
				  <select name="depth"
					  id="depth"
					  size="1"
					  title="検索階層数: 数字を増やすと検索速度は落ちますがマッチする可能性が上がります。">
					  %s
				  </select>
				  <input type="submit" name="submit" value="検索">
			    </form>
				<table>`, s, d, d, s, ao, de)
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

// andorPadding : 検索ワードのスペースをandなら".*" orなら"|"で埋める
func andorPadding(s, method string) string {
	ss := strings.Fields(s)
	if method == "and" {
		method = ".*"
		s = strings.Join(ss, method)
	} else if method == "or" {
		method = "|"
		s = strings.Join(ss, method)
		s = "(" + s + ")"
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
	opt = append(opt, andorPadding(receiveValue, searchAndOr))
	opt = append(opt, slashedDirPath)
	// opt = append(opt, "2>", "/dev/null")
	fmt.Println(opt)
	out, err := exec.Command("rga", opt...).Output()
	if err != nil {
		fmt.Println(err)
	}
	// 結果をarray型に格納
	outstr := string(out)
	fmt.Println(outstr)
	results := strings.Split(outstr, "\n")
	results = results[:len(results)-1] // Pop last element cause \\n

	/* html表示 */
	// 検索後のフォームに再度同じキーワードを入力
	fmt.Fprintf(w, htmlClause(receiveValue, directoryPath,
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
		func() string {
			s := `<input type="radio" value="and" name="andor-search">and
				 <input type="radio" value="or"  name="andor-search">or`
			return strings.Replace(s,
				"\"andor-search\">"+searchAndOr,
				"\"andor-search\"checked=\"checked\">"+searchAndOr,
				1)
		}(),
	))
	match := regexp.MustCompile(`^\d`)
	for _, s := range results {
		if match.MatchString(s) { // 行数から始まるときはfile contents
			fmt.Fprintf(w, // => http.ResponseWriter
				`<tr> <td> %s </td> <tr>`, highlightString(
					html.EscapeString(s),
					// メタ文字含まない検索文字のみhighlight
					strings.Fields(receiveValue)...),
			)
		} else { // 行数から始まらないときはfile name
			fmt.Fprintf(w, `<tr> <td> %s </td> <tr>`, highlightFilename(s))
		}
	}
	fmt.Fprintln(w, `</table>
				</body>
				</html>`)
}

// ファイル名をリンク化したhtmlを返す
func highlightFilename(s string) string {
	dirpath := filepath.Dir(s)

	// drive path convert
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

func main() {
	flag.BoolVar(&showVersion, "v", false, "show version")
	flag.BoolVar(&showVersion, "version", false, "show version")
	flag.Parse()
	if showVersion {
		fmt.Println("grep-server", VERSION)
		return // versionを表示して終了
	}

	http.HandleFunc("/", showInit)           // top page
	http.HandleFunc("/searching", addResult) // search result
	http.ListenAndServe(":8080", nil)
}
