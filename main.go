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
	// CAP : 表示する検索結果上限数
	CAP = 1000
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

func htmlClause(s, d string) string {
	return fmt.Sprintf(
		`<!DOCTYPE html>
			<html>
			<head>
			<meta http-equiv="Content-Type" content="text/html; charaset=utf-8">
			<title>Grep Server %s %s</title>
			</head>
			  <body>
			    <form method="get" action="/searching">
				<input type="text" placeholder="フォルダパス(ex:/usr/bin ex:\ShareUsers\User\Personal)" name="directory-path" id="directory-path" value="%s" size="50">
				  <a href=https://github.com/u1and0/grep-server/blob/master/README.md>Help</a>
				  <br>
				  <input type="text" placeholder="検索語" name="query" value="%s" size="50">
				  <input type="submit" name="submit" value="検索">
			    </form>
				<table>`, s, d, d, s)
}

//old input <input type="file" name="directory-path" value="Browse" webkitdirectory />

// Top page
func showInit(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, htmlClause("", ""))
}

func addResult(w http.ResponseWriter, r *http.Request) {
	// Modify query
	receiveValue := r.FormValue("query")
	directoryPath := r.FormValue("directory-path")
	commandDir := directoryPath
	if *root != "" {
		commandDir = strings.TrimPrefix(commandDir, *root)
	}
	if *pathSplitWin {
		// filepath.ToSlash(directoryPath) <= Windows版Goでしか有効でない
		commandDir = strings.ReplaceAll(commandDir, `\`, "/")
	}

	// コマンド生成
	// searchWord := []string{"会社", "生産", "弁当"}
	// `rga -n --no-heading "search word"` と同じ動き
	opt := []string{ // rga/rg options
		"--line-number",
		"--max-columns", "160",
		"--max-columns-preview",
		"--heading",
	}
	opt = append(opt, receiveValue) // search words
	opt = append(opt, commandDir)   // directory path
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
	fmt.Fprintf(w, htmlClause(receiveValue, directoryPath))
	match := regexp.MustCompile(`^\d`)
	for _, s := range results {
		if match.MatchString(s) { // 行数から始まるとき
			fmt.Fprintf(w, `<tr> <td> %s </td> <tr>`, highlightString(html.EscapeString(s), receiveValue))
		} else {
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
		// dirpath = strings.ReplaceAll(dirpath, "/", `\`)
	}

	if s != "" {
		s = strings.Replace(s, s,
			"<a href=\"file://"+s+"\">"+s+"</a>", 1)
		s += " <a href=\"file://" + dirpath + "\" title=\"<< クリックでフォルダに移動\"><<</a>"
	}
	return s
}

// sの文字列中にあるwordsの背景を黄色にハイライトしたhtmlを返す
func highlightString(s string, words ...string) string {
	for _, w := range words {
		re := regexp.MustCompile(`((?i)` + w + `)`)
		found := re.FindString(s)
		if found != "" {
			s = strings.Replace(s, found, "<span style=\"background-color:#FFCC00;\">"+found+"</span>", 1)
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

	http.HandleFunc("/", showInit)
	http.HandleFunc("/searching", addResult)
	http.ListenAndServe(":8080", nil)
}
