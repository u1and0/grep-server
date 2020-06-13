package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	cmd "grep-server/cmd"
)

const (
	// VERSION : version
	VERSION = "1.0.3"
	// EXE : Search command
	EXE = "/usr/bin/rga"
	// LOGFILE : 検索条件 / マッチファイル数 / マッチ行数 / 検索時間を記録するファイル
	LOGFILE = "/var/log/grep-server.log"
	// PORT : http.ListenAndServe port number
	PORT = ":8080"
)

var (
	showVersion  bool
	debug        bool
	root         string
	encoding     string
	pathSplitWin bool
	timeout      time.Duration
)

func main() {
	// Command check
	if _, err := exec.LookPath(EXE); err != nil {
		log.Fatalf("[ERROR] %s", err.Error())
	}
	// Parse flags
	flag.BoolVar(&showVersion, "v", false, "show version")
	flag.BoolVar(&showVersion, "version", false, "show version")
	flag.BoolVar(&debug, "debug", false, "run as debug mode")
	flag.StringVar(&root, "r", "", "Append root directory path")
	flag.StringVar(&root, "root", "", "Append root directory path")
	flag.StringVar(&encoding, "E", "UTF-8", "Set default encoding")
	flag.StringVar(&encoding, "encoding", "UTF-8", "Set default encoding")
	flag.BoolVar(&pathSplitWin, "s", false, "OS path split windows backslash")
	flag.BoolVar(&pathSplitWin, "sep", false, "OS path split windows backslash")
	flag.DurationVar(&timeout, "t", 10*time.Second, "Search method timeout")
	flag.DurationVar(&timeout, "timeout", 10*time.Second, "Search method timeout")
	flag.Parse()
	// Show version
	if showVersion {
		fmt.Println("grep-server", VERSION)
		rgaVersion, _ := exec.Command(EXE, "--version").Output()
		fmt.Println(string(rgaVersion))
		return // versionを表示して終了
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
	http.ListenAndServe(PORT, nil)
}

// showInit : Top page html
func showInit(w http.ResponseWriter, r *http.Request) {
	// 検索語、ディレクトリは空
	// 検索階層は何もselectされていない(デフォルトは一番上の1になる)
	s := cmd.Search{Depth: "1", AndOr: "and", Encoding: encoding}
	if debug {
		fmt.Printf("[DEBUG] search struct: %+v\n", s)
	}
	fmt.Fprintf(w, s.HTMLClause())
}

// addResult : Print ripgrep-all result as html contents
func addResult(w http.ResponseWriter, r *http.Request) {
	// Modify query
	search := cmd.Search{
		Keyword:      r.FormValue("query"),
		Path:         r.FormValue("directory-path"),
		AndOr:        r.FormValue("andor-search"),
		Depth:        r.FormValue("depth"),
		Encoding:     r.FormValue("encoding"),
		CmdKeyword:   "",
		CmdPath:      r.FormValue("directory-path"), // 初期値はPathと同じ
		Exe:          EXE,
		Root:         root,
		PathSplitWin: pathSplitWin,
		Debug:        debug,
		Timeout:      timeout,
	}
	if debug {
		fmt.Printf("[DEBUG] search struct: %+v\n", search)
	}

	/* html表示 */
	fmt.Fprintf(w, search.HTMLClause())     // 検索後のフォームに再度同じキーワードを入力
	defer fmt.Fprintf(w, `</body> </html>`) // 終了タグ
	/* 検索結果表示 */
	outstr, err := search.Grep()
	if err != nil { // Error
		fmt.Fprintf(w, `<h4> %s </h4>`, err)
		log.Printf(
			"[ERROR] %s Keyword: [ %-30s ] Path: [ %-50s ]\n",
			err, search.Keyword, search.Path)
	} else { // Success
		result := cmd.Result{Out: outstr, Root: root, PathSplitWin: pathSplitWin}
		result = result.HTMLContents(search.Keyword)
		if debug {
			fmt.Printf("[DEBUG] result struct: %+v\n", result)
		}
		fmt.Fprintf(w, "<h4>")
		for _, h := range result.Stats {
			fmt.Fprintf(w, "%s<br>", h)
		}
		fmt.Fprintf(w, "</h4>")
		fmt.Fprintf(w, `<table>`)
		for _, h := range result.Contents {
			fmt.Fprintf(w, `<tr> <td>%s</td> </tr>`, h)
		}
		fmt.Fprintf(w, `</table>`)
		log.Printf(
			"%s Keyword: [ %-30s ] Path: [ %-50s ]\n",
			strings.Join(result.Stats, " "), search.Keyword, search.Path)
	}
}
