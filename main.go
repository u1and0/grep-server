package main

import (
	"fmt"
	"html"
	"net/http"
	"os/exec"
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
			<title>Grep Server</title>
			</head>
			<body>
			%s
			<table>`, s)
}

func process(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, htmlClause(""))
	// コマンド生成
	searchWord := "file"
	// `rg -n --no-heading "search word"` と同じ動き
	out, err := exec.Command("rg", "-n", searchWord).Output()
	if err != nil {
		fmt.Println(err)
	}
	// 結果をarray型に格納
	outstr := string(out)
	fmt.Println(outstr)
	results := strings.Split(outstr, "\n")
	results = results[:len(results)-1] // Pop last element cause \\n

	var pm []PathMap
	for _, r := range results {
		spl := strings.SplitN(r, ":", -1)
		fmt.Println(spl)
		pm = append(pm, PathMap{File: spl[0], Line: spl[1], Highlight: spl[2]}) //strings.Join(spl[2:], "")})
	}
	fmt.Println(pm)
	for _, p := range pm {
		fmt.Fprintf(w,
			`
			<tr>
				<td> %s </td>
				<td> %s </td>
				<td> %s </td>
			</tr>
		`, highlightFilename(p.File, []string{".+"}),
			p.Line,
			highlightString(html.EscapeString(p.Highlight), []string{searchWord}), // 検索ワードハイライト
		)
	}

	// templateにHTML渡すのは非array形式、go側でHTML表示に整形する
	// var htm []string
	// for _, r := range hiresults {
	// 	htm = append(htm, "<tr><td>"+r+"</td></tr>")
	// }
	// html := strings.Join(htm, "\n")

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
func highlightFilename(s string, words []string) string {
	for _, w := range words {
		re := regexp.MustCompile(`((?i)` + w + `)`)
		/* Replace only a word
		全て変える re.ReplaceAll(s, "<span style=\"background-color:#FFCC00;\">$1</span>")　は削除 */
		found := re.FindString(s)
		if found != "" {
			s = strings.Replace(s, found,
				"<a href=\"file:///home/u1and0/Dropbox/Program/go/src/github.com/u1and0/grep-server/"+found+"\">"+found+"</a>", 1)
		}
	}
	return s
}

// sの文字列中にあるwordsの背景を黄色にハイライトしたhtmlを返す
func highlightString(s string, words []string) string {
	for _, w := range words {
		re := regexp.MustCompile(`((?i)` + w + `)`)
		/* Replace only a word
		全て変える re.ReplaceAll(s, "<span style=\"background-color:#FFCC00;\">$1</span>")　は削除 */
		found := re.FindString(s)
		if found != "" {
			s = strings.Replace(s, found, "<span style=\"background-color:#FFCC00;\">"+found+"</span>", 1)
		}
	}
	return s
}

func main() {
	server := http.Server{
		Addr: ":8080",
	}
	http.HandleFunc("/process", process)
	server.ListenAndServe()
}
