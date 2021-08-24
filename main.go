package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	cmd "grep-server/cmd"

	"github.com/op/go-logging"
)

const (
	// VERSION : version
	VERSION = "2.0.0"
	// EXE : Search command ripgrep-all
	EXE = "/usr/bin/rga"
	// EXF : Search command ripgrep
	EXF = "/usr/bin/rg"
	// LOGFILE : 検索条件 / マッチファイル数 / マッチ行数 / 検索時間を記録するファイル
	LOGFILE = "/var/log/grep-server.log"
)

var (
	showVersion  bool
	debug        bool
	root         string
	trim         string
	encoding     string
	pathSplitWin bool
	timeout      time.Duration
	port         int
	log          = logging.MustGetLogger("main")
)

func main() {
	// Parse flags
	flag.BoolVar(&showVersion, "v", false, "show version")
	flag.BoolVar(&showVersion, "version", false, "show version")
	flag.BoolVar(&debug, "debug", false, "run as debug mode")
	flag.StringVar(&root, "r", "", "Append root directory path")
	flag.StringVar(&root, "root", "", "Append root directory path")
	flag.StringVar(&trim, "T", "", "DB trim prefix for directory path")
	flag.StringVar(&trim, "trim", "", "DB trim prefix for directory path")
	flag.StringVar(&encoding, "E", "UTF-8", "Set default encoding")
	flag.StringVar(&encoding, "encoding", "UTF-8", "Set default encoding")
	flag.BoolVar(&pathSplitWin, "s", false, "OS path split windows backslash")
	flag.BoolVar(&pathSplitWin, "sep", false, "OS path split windows backslash")
	flag.DurationVar(&timeout, "t", 10*time.Second, "Search method timeout")
	flag.DurationVar(&timeout, "timeout", 10*time.Second, "Search method timeout")
	flag.IntVar(&port, "port", 8080, "http.ListenAndServe port number. Default access to http://localhost:8080/")
	flag.IntVar(&port, "p", 8080, "http.ListenAndServe port number. Default access to http://localhost:8080/")
	flag.Parse()
	// Show version
	if showVersion {
		fmt.Println("grep-server", VERSION)
		rgaVersion, _ := exec.Command(EXE, "--version").Output()
		rgVersion, _ := exec.Command(EXF, "--version").Output()
		fmt.Println(string(rgaVersion), string(rgVersion))
		return // versionを表示して終了
	}
	// Log setting
	logfile, err := os.OpenFile(LOGFILE, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	defer logfile.Close()
	setLogger(logfile) // log.XXX()を使うものはここより後に書く
	if err != nil {
		log.Panicf("%s", err.Error())
	}

	// Command check
	if _, err := exec.LookPath(EXE); err != nil {
		log.Panicf("%s", err.Error())
	}

	// HTTP response
	http.HandleFunc("/", showInit)        // top page
	http.HandleFunc("/search", addResult) // search result
	pt := ":" + strconv.Itoa(port)        // => :8080
	log.Infof("Server open.")
	http.ListenAndServe(pt, nil)
}

// setLogger is printing out log message to STDOUT and LOGFILE
func setLogger(f *os.File) {
	var format = logging.MustStringFormatter(
		`%{color}[%{level:.6s}] ▶ %{time:2006-01-02 15:04:05} %{shortfile} %{message} %{color:reset}`,
	)
	backend1 := logging.NewLogBackend(os.Stdout, "", 0)
	backend2 := logging.NewLogBackend(f, "", 0)
	backend1Formatter := logging.NewBackendFormatter(backend1, format)
	backend2Formatter := logging.NewBackendFormatter(backend2, format)
	logging.SetBackend(backend1Formatter, backend2Formatter)
}

// showInit : Top page html
func showInit(w http.ResponseWriter, r *http.Request) {
	// オプションのデフォルト設定
	// 検索語、ディレクトリは空
	// 検索階層は何もselectされていない(デフォルトは一番上の1になる)
	s := cmd.Search{Depth: "1", AndOr: "and", Encoding: encoding, Mode: "Content"}
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
		Mode:         r.FormValue("mode"),
		CmdKeyword:   "",
		CmdPath:      r.FormValue("directory-path"), // 初期値はPathと同じ
		Exe:          EXE,
		Exf:          EXF,
		Root:         root,
		Trim:         trim, // Path prefix trim
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

	/* コマンド作成 */
	c, err := search.CommandGen()
	if debug {
		fmt.Printf("[DEBUG] options: %v\n", c)
	}
	if err != nil { // Error
		fmt.Fprintf(w, `<h4> %s </h4>`, err)
		log.Errorf(
			"%s Keyword: [ %-30s ] Path: [ %-50s ]\n",
			err, search.Keyword, search.Path)
	}

	/* 検索結果表示 */
	outstr, err := search.Grep(c)
	if debug {
		fmt.Printf("[DEBUG] result: %+v\n", outstr)
	}
	if fmt.Sprintf("%s", err) == "exit status 1" {
		fmt.Fprintf(w, `<h4> %s </h4>`, "検索がマッチしませんでした。")
		log.Errorf(
			"[ERROR] %s Keyword: [ %-30s ] Path: [ %-50s ]\n",
			err, search.Keyword, search.Path)
	} else if err != nil { // Error
		fmt.Fprintf(w, `<h4> %s </h4>`, err)
		log.Errorf(
			"[ERROR] %s Keyword: [ %-30s ] Path: [ %-50s ]\n",
			err, search.Keyword, search.Path)
	} else { // Success
		result := cmd.Result{Out: outstr, Root: root, Trim: trim, PathSplitWin: pathSplitWin}
		ss := strings.Fields(search.Keyword)
		result = result.HTMLContents(ss)
		if debug {
			log.Debugf("result struct: %+v\n", result)
		}
		fmt.Fprintf(w, "<h4>")
		// Stats 出力
		for _, h := range result.Stats {
			fmt.Fprintf(w, "%s<br>", h)
		}
		fmt.Fprintf(w, "</h4>")
		// 検索結果出力
		fmt.Fprintf(w, `<table>`)
		for _, h := range result.Contents {
			if h.File != "" {
				fmt.Fprintf(w, `<tr> <td>%s %s</td> </tr>`, h.File, h.Dir)
			} else {
				fmt.Fprintf(w, `<tr> <td>%s</td> </tr>`, h.Highlight)
			}
		}
		fmt.Fprintf(w, `</table>`)
		log.Noticef(
			"%s Keyword: [ %-30s ] Path: [ %-50s ]\n",
			strings.Join(result.Stats, " "), search.Keyword, search.Path)
	}
}
