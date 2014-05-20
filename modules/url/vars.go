package url

import (
	"regexp"

	"github.com/crimsonvoid/irclib/module"
)

const (
	maxContentLen = 150
)

var (
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
	ytVidAPI    = "https://gdata.youtube.com/feeds/api/videos/%v?v=2&alt=jsonc"
	ytVidRegexp = regexp.MustCompile(`youtu(be\.com/watch\?v=|\.be/)(?P<id>[\w_\-]{11})`)
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
	ytPLAPI    = "https://gdata.youtube.com/feeds/api/playlists/%v?v=2&alt=jsonc"
	ytPLRegexp = regexp.MustCompile(`youtube\.com/playlist\?list=(?P<id>(PL)?[\w_\-]+)`)
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
type fourChJSON struct {
	Posts []struct {
		Sub, Com string
		No       int
	}
}

var (
	fourChRegexp = regexp.MustCompile(`4chan\.org/(?P<board>[[:alnum:]]+)/thread/(?P<thread>\d+)(#p(?P<post>\d+))?`)
	fourChAPI    = "https://a.4cdn.org/%v/thread/%v.json"
)

// Reddit

// Soundcloud

// Twitter

// Imgur

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
}
