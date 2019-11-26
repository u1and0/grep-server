package main

import (
	"fmt"
	"html"
	"net/http"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// PathMap : File:ファイルネームを起点として、
// そのディレクトリと検索語をハイライトした文字列を入れる
type PathMap struct {
	File      string
	Line      string
	Dir       string
	Highlight string
}

func htmlClause(s string) string {
	return fmt.Sprintf(
		`<!DOCTYPE html>
			<html>
			<head>
			<meta http-equiv="Content-Type" content="text/html; charaset=utf-8">
			<title>Grep Server %s</title>
			</head>
			  <body>
			    <form method="get" action="/searching">
  				  <input type="file" name="path-select" id="path-select" value="Browse" webkitdirectory />
				  <br>
				  <input type="text" name="query" value="%s" size="50">
				  <input type="submit" name="submit" value="検索">
				  <a href=https://github.com/u1and0/locate-server/blob/master/README.md>Help</a>
			    </form>
				<table>`, s, s)
}

// Top page
func showInit(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, htmlClause(""))
}

func addResult(w http.ResponseWriter, r *http.Request) {
	// Modify query
	receiveValue := r.FormValue("query")

	// コマンド生成
	// searchWord := []string{"会社", "生産", "弁当"}
	// `rga -n --no-heading "search word"` と同じ動き
	out, err := exec.Command("rga", "-n", "--heading", receiveValue, "/home/vagrant/Dropbox/Document/k会社", "2>", "/dev/null").Output()
	if err != nil {
		fmt.Println(err)
	}
	// 結果をarray型に格納
	outstr := string(out)
	fmt.Println(outstr)
	results := strings.Split(outstr, "\n")
	results = results[:len(results)-1] // Pop last element cause \\n

	// html表示
	fmt.Fprintf(w, htmlClause(receiveValue))
	match := regexp.MustCompile(`^\d`)
	for _, r := range results {
		if match.MatchString(r) {
			fmt.Fprintf(w, `<tr> <td> %s </td> <tr>`, highlightString(html.EscapeString(r), receiveValue))
		} else {
			fmt.Fprintf(w, `<tr> <td> %s </td> <tr>`, highlightFilename(r))
		}
	}

	// template読込
	// fmt.Println(html)
	// t, _ := template.ParseFiles("range.html") // 結果を書き込むhtmlファイル
	// t.Execute(w, template.HTML(r.FormValue(html)))

	// template読み込まない方式
	fmt.Fprintln(w, `</table>
				</body>
				</html>`)
}

// sの文字列中にあるwordsを紫色に変えてリンク化したhtmlを返す
func highlightFilename(s string) string {
	re := regexp.MustCompile(`((?i)` + "^/.*" + `)`) // /から始まる全ての文字列
	found := re.FindString(s)
	dirpath := filepath.Dir(found)
	if found != "" {
		s = strings.Replace(s, found,
			"<a href=\"file://"+found+"\">"+found+"</a>", 1)
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
	http.HandleFunc("/", showInit)
	http.HandleFunc("/searching", addResult)
	http.ListenAndServe(":8080", nil)
}
