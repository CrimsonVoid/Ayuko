package url

import (
	"fmt"
	"regexp"

	"github.com/crimsonvoid/irclib/module"
)

const (
	maxContentLen = 100
)

var (
	htmlCleanerR = regexp.MustCompile(fmt.Sprintf(`</?[%v].*?>`,
		`a|br|code|span|wbr`,
	))

	urlRe = regexp.MustCompile(`(http|https)\://[a-zA-Z0-9\-\.]+\.[a-zA-Z]{2,3}(:[a-zA-Z0-9]*)?/?([a-zA-Z0-9\-\._\?\,\'/\\\+&amp;%\$#\=~])*`)

	Module *module.Module
)

// YouTube - Video
type ytVidJSON struct {
	Data struct {
		Id       string
		Uploader string
		Title    string
		Duration int
	}
}

var (
	ytVidRegexp = regexp.MustCompile(`youtu(be\.com/watch\?v=|\.be/)(?P<id>[\w_\-]{11})`)
	ytVidAPI    = "https://gdata.youtube.com/feeds/api/videos/%v?v=2&alt=jsonc"
)

// Youtube - Playlist
type ytPLJSON struct {
	Data struct {
		Id         string
		Author     string
		Title      string
		TotalItems int
	}
}

var (
	ytPLRegexp = regexp.MustCompile(`youtube\.com/playlist\?list=(?P<id>(PL)?[\w_\-]+)`)
	ytPLAPI    = "https://gdata.youtube.com/feeds/api/playlists/%v?v=2&alt=jsonc"
)

type githubRepoJSON struct {
	Html_url    string
	Description string
	Language    string
	Homepage    string
}

type githubIssueJSON struct {
	Html_url string
	Title    string
	State    string
}

type githubCommitJSON struct {
	Html_url string
	Commit   struct {
		Message string
	}
}

var (
	githubRegexp = regexp.MustCompile(
		`github\.com/(?P<user>.*?)/(?P<repo>.*?)($|/(?P<extra>.*))`)
	githubIORegexp = regexp.MustCompile(`(http(s)?://)?(?P<user>.*)\.github\.io/(?P<repo>.*?)($|/)`)
	githubAPI      = "https://api.github.com/repos/%v/%v"
)

// Vimeo
type vimeoJSON struct {
	Id       int    `json:"id"`
	Title    string `json:"title"`
	Url      string `json:"url"`
	Username string `json:"user_name"`
	Duration int    `json:"duration"`
}

var (
	vimeoRegexp = regexp.MustCompile(`vimeo\.com/(?P<id>\d{8})`)
	vimeoAPI    = "https://vimeo.com/api/v2/video/%v.json"
)

var (
	steamRegexp = regexp.MustCompile(`store\.steampowered\.com/app/(?P<id>\d*)`)
	hnRegexp    = regexp.MustCompile(`news\.ycombinator\.com/item\?id=(?P<id>\d*)`)
)

// Reddit
// Soundcloud
// Twitter

/*
Input:
	url to get

Return:
	string to print
	error from parsing url
*/
var parseMap = []struct {
	re *regexp.Regexp
	fn func(*regexp.Regexp, string) (string, error)
}{
	{ytVidRegexp, ytVidParser},
	{ytPLRegexp, ytPLParser},
	{githubRegexp, githubParser},
	{githubIORegexp, githubParser},
	{vimeoRegexp, vimeoParser},
	{steamRegexp, steamParser},
	{hnRegexp, hnParser},
}
