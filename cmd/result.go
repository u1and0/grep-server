package cmd

import (
	"path/filepath"
	"regexp"
	"strings"
)

const (
	// STATSLENGTH : rg --stats の行数
	STATSLENGTH = 8
)

// Result : rga結果, Statsと結果に別れる
type Result struct {
	Out          []string  // rgaの結果の生の文字列
	Stats        []string  // rga --stat の結果
	Contents     []Content // rgaの結果をHTML装飾した文字列
	Root, Trim   string
	PathSplitWin bool // Trueでスラッシュをバックスラッシュに変えるフラグ
}

// Content :
type Content struct{ Dir, File, Highlight string }

// ファイル名をリンク化したhtmlを返す
func (r *Result) highlightFilename(s string, words []string) (string, string) {
	dirpath := filepath.Dir(s)
	if r.Trim != "" { // Trim drive path
		s = strings.TrimPrefix(s, r.Trim)
		dirpath = strings.TrimPrefix(dirpath, r.Trim)
	}
	if r.Root != "" && s != "" { // Add drive path
		s = r.Root + s
		dirpath = r.Root + dirpath
	}
	if r.PathSplitWin { // Windows path convert
		s = strings.ReplaceAll(s, "/", `\`)
	}
	if s != "" {
		s = strings.Replace(s, s,
			"<a target=\"_blank\" href=\"file://"+s+"\">"+highlightString(s, words)+"</a>", 1)
		dirpath = "<a href=\"file://" + dirpath + "\" title=\"<< クリックでフォルダに移動\"><<</a>"
	}
	return s, dirpath
}

// highlightString : sの文字列中にあるwordsの背景を黄色にハイライトしたhtmlを返す
func highlightString(s string, words []string) string {
	for _, w := range words {
		re := regexp.MustCompile(`((?i)` + w + `)`) // ((?i)word)
		found := re.FindString(s)
		color := "style=\"background-color:#FFCC00;\">"
		if found != "" {
			s = strings.ReplaceAll(s, found, "<span "+color+found+"</span>")
		}
	}
	return s
}

// HTMLContents : ファイル名ハイライトとファイルコンテンツハイライトを
// Result構造体に入れて返す
func (r *Result) HTMLContents(words []string) Result {
	var (
		l = len(r.Out) - STATSLENGTH
		x = regexp.MustCompile(`^/`)
	)
	for _, s := range r.Out[:l] {
		var c Content
		if x.MatchString(s) { // '/'から始まるときはfilename
			c.File, c.Dir = r.highlightFilename(s, words)
		} else { // '/'から始まらないときはfile contents
			c.Highlight = highlightString(s, words)
		}
		r.Contents = append(r.Contents, c)
	}
	r.Stats = r.Out[l:] // 最後の8行はrga --stats の統計情報
	return *r
}
