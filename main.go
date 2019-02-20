package main

import (
	"fmt"
	"github.com/ChimeraCoder/anaconda"
	"github.com/snow-dd51/shigoowa/conf"
	"net/url"
	"regexp"
	"time"
)

const (
	timeStampFormat = "2006-01-02 15:04:05"
	confPath        = "./auth.json"
)

func main() {
	app, err := NewApp(confPath)
	if err != nil {
		return
	}
	app.mainLoop()
}

type App struct {
	IsDebug      bool
	SleepSeconds int
	TwAPI        *anaconda.TwitterApi
	Conf         *conf.AppConf
	TwProc       TweetProcessor
}

type TweetProcessor interface {
	Match(anaconda.Tweet) bool
	Make(anaconda.Tweet) string
}

func Debugf(msg string, arg ...interface{}) {
	a := make([]interface{}, 1)
	a[0] = time.Now().Format(timeStampFormat)
	a = append(a, arg...)
	fmt.Printf("[%s] "+msg+"\n", a...)
}

func NewApp(confpath string) (*App, error) {
	ac := conf.NewAppConf()
	err := ac.Read(confPath)
	if err != nil {
		return nil, err
	}
	var debug bool
	if !ac.IsProd {
		debug = true
		Debugf("Debug Mode!")
	}
	sleepSeconds := 61
	if ac.SleepSeconds > 60 {
		sleepSeconds = ac.SleepSeconds
	}
	var proc TweetProcessor
	if debug {
		proc = DevProc{}
	} else {
		proc = DefaultProc{}
	}
	api := anaconda.NewTwitterApiWithCredentials(
		ac.AccessToken,
		ac.AccessSecret,
		ac.ConsumerKey,
		ac.ConsumerSecret,
	)
	u, err := api.GetSelf(nil)
	if err != nil {
		return nil, err
	}
	Debugf("authorized as %s\n", u.ScreenName)
	return &App{
		IsDebug:      debug,
		SleepSeconds: sleepSeconds,
		TwAPI:        api,
		Conf:         ac,
		TwProc:       proc,
	}, nil
}

func (app *App) mainLoop() {
	if !app.validateConf() {
		return
	}
	ln := 0
	lastStatusId := ""
	if app.IsDebug {
		lastStatusId = ""
		Debugf("last status ID is %s", lastStatusId)
	} else {
		// 再起動に備えて開始地点は保存したい
		lastStatusId = app.Conf.LastStatusID
	}
	for true {
		reqv := url.Values{}
		if lastStatusId != "" {
			reqv.Add("since_id", lastStatusId)
		}
		reqv.Add("count", "20")
		Debugf("Loop %d", ln)
		ln++
		tl, err := app.TwAPI.GetHomeTimeline(reqv)
		if err != nil {
			Debugf("%v", err)
			break
		}
		for i, v := range tl {
			// v.Textは<文字数>byteで切られているので使ってはいけない罠
			fmt.Printf("%d: %s\n%s\n%s\n", i, v.User.ScreenName, v.FullText, v.CreatedAt)
			if app.TwProc.Match(v) {
				newTweet := app.TwProc.Make(v)
				if newTweet != "" {
					opt := url.Values{}
					app.TwAPI.PostTweet(newTweet, opt)
				} else {
					Debugf("match")
				}
			}
		}
		if len(tl) > 0 {
			lastStatusId = tl[0].IdStr
			app.Conf.LastStatusID = lastStatusId
			app.Conf.Write(confPath)
			Debugf("%v", lastStatusId)
		}
		if lastStatusId == "" {
			break
		}
		time.Sleep(time.Duration(app.SleepSeconds) * time.Second)
	}
}

type DefaultProc struct{}

func (p DefaultProc) Match(tw anaconda.Tweet) bool {
	m, err := regexp.MatchString("^しごおわ$", tw.FullText)
	if err != nil {
		return false
	}
	return m
}
func (p DefaultProc) Make(tw anaconda.Tweet) string {
	return "今日も一日お仕事お疲れさま♡毎日頑張って偉いね！"
}

type DevProc struct{}

func (p DevProc) Match(tw anaconda.Tweet) bool {
	m, err := regexp.MatchString("アルストロメリア", tw.FullText)
	if err != nil {
		return false
	}
	return m
}
func (p DevProc) Make(tw anaconda.Tweet) string {
	return "やるぞ " + time.Now().Format(timeStampFormat)
}

func (app *App) validateConf() bool {
	if app.SleepSeconds < 0 {
		return false
	}
	return true
}
