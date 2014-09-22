package url

import (
	"fmt"
	"regexp"

	"github.com/crimsonvoid/irclib/module"
)

const (
	maxContentLen = 80
)

var (
	htmlCleanerR = regexp.MustCompile(fmt.Sprintf(`</?[%v].*?>`,
		`a|br|code|span`,
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

// Github
type githubJSON struct {
	Html_url    string
	Description string
	Language    string
	Homepage    string
}

var (
	githubRegexp   = regexp.MustCompile(`github.com/(?P<user>.*?)/(?P<repo>.*?)($|/)`)
	githubIORegexp = regexp.MustCompile(`(http(s)?://)?(?P<user>.*)\.github.io/(?P<repo>.*?)($|/)`)
	githubAPI      = "https://api.github.com/repos/%v/%v"
)

// 4Chan
type chanJSON struct {
	Posts []struct {
		Sub, Com string
		No       int
	}
}

var (
	fourChRegexp = regexp.MustCompile(`4chan\.org/(?P<board>[[:alnum:]]+)/thread/(?P<thread>\d+)(#(?P<rel>p)(?P<post>\d+))?`)
	fourChAPI    = "https://a.4cdn.org/%v/thread/%v.json"
	fourChFmt    = "[https://boards.4chan.org/%v/thread/%v%v] %v%v"

	eightChRegexp = regexp.MustCompile(`8chan\.co/(?P<board>[[:alnum:]]+)/res/(?P<thread>\d+)\.html(#(?P<rel>)(?P<post>\d+))?`)
	eightChAPI    = "https://8chan.co/%v/res/%v.json"
	eightChFmt    = "[https://8chan.co/%v/res/%v.html%v] %v%v"
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
	vimeoRegexp = regexp.MustCompile(`vimeo.com/(?P<id>\d{8})`)
	vimeoAPI    = "https://vimeo.com/api/v2/video/%v.json"
)

var (
	steamRegexp = regexp.MustCompile(`store.steampowered.com/app/(?P<id>\d*)`)
	hnRegexp    = regexp.MustCompile(`news.ycombinator.com/item?id=(?P<id>\d*)`)
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
	{fourChRegexp, fourChParser},
	{eightChRegexp, eightChParser},
	{vimeoRegexp, vimeoParser},
	{steamRegexp, steamParser},
	{hnRegexp, hnParser},
}
