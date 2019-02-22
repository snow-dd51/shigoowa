package main

import (
	"flag"
	"fmt"
	"github.com/ChimeraCoder/anaconda"
	"github.com/snow-dd51/shigoowa/conf"
	"net/url"
	"regexp"
	"time"
)

const (
	// Userからとれないので残念ながら決め打ち
	tweetTimeZone   = "UTC+9"
	tweetTimeOffset = 9 * 60 * 60
	timeStampFormat = "2006-01-02 15:04:05"
)

var (
	confPathp = flag.String("conf", "./auth.json", "config file path")
	debugp    = flag.Bool("debug", false, "debug mode")
)

func main() {
	flag.Parse()
	app, err := NewApp(*confPathp)
	if err != nil {
		fmt.Printf("NewApp Error: %v", err)
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
	MyInfo       anaconda.User
	ConfigPath   string
}

type TweetProcessor interface {
	Match(anaconda.Tweet) (bool, string)
}

func Debugf(msg string, arg ...interface{}) {
	a := make([]interface{}, 1)
	a[0] = time.Now().Format(timeStampFormat)
	a = append(a, arg...)
	fmt.Printf("[%s] "+msg+"\n", a...)
}

func NewApp(confpath string) (*App, error) {
	ac := conf.NewAppConf()
	err := ac.Read(confpath)
	if err != nil {
		return nil, err
	}
	var debug bool
	if !ac.IsProd {
		debug = true
		Debugf("Debug Mode!")
	} else if *debugp {
		debug = true
		Debugf("Debug Mode by flag!")
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
	api.HttpClient.Timeout, err = time.ParseDuration("10s")
	if err != nil {
		return nil, err
	}
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
		MyInfo:       u,
		ConfigPath:   confpath,
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
			if v.InReplyToStatusIdStr != "" ||
				v.InReplyToUserIdStr != "" {
				Debugf("!!! Skip, Reply !!!")
				continue
			}
			if v.User.Id == app.MyInfo.Id {
				Debugf("!!! Skip, Mine !!!")
				continue
			}
			if v.RetweetedStatus != nil {
				Debugf("%d: %s", i, v.User.ScreenName)
				Debugf("!!! Skip, Retweeted !!!")
				continue
			} else {
				fmt.Printf("%d: %s\n%s\n%s\n", i, v.User.ScreenName, v.FullText, v.CreatedAt)
			}
			tt, err := v.CreatedAtTime()
			if err != nil {
				Debugf("tterr %v", err)
			} else {
				Debugf("%s", inJST(tt).Format(timeStampFormat))
			}
			if m, newTweet := app.TwProc.Match(v); m {
				if newTweet != "" {
					opt := url.Values{}
					opt.Add("in_reply_to_status_id", v.IdStr)
					prefix := fmt.Sprintf("@%s ", v.User.ScreenName)
					app.TwAPI.PostTweet(prefix+newTweet, opt)
				} else {
					Debugf("match")
				}
			}
		}
		if len(tl) > 0 {
			lastStatusId = tl[0].IdStr
			app.Conf.LastStatusID = lastStatusId
			app.Conf.Write(app.ConfigPath)
			Debugf("%v", lastStatusId)
		}
		if lastStatusId == "" {
			break
		}
		time.Sleep(time.Duration(app.SleepSeconds) * time.Second)
	}
}

func inJST(t0 time.Time) time.Time {
	tzloc := time.FixedZone(tweetTimeZone, tweetTimeOffset)
	return t0.In(tzloc)
}

type DefaultProc struct{}

func (p DefaultProc) Match(tw anaconda.Tweet) (bool, string) {
	m, _ := regexp.MatchString("^しごおわ$", tw.FullText)
	if m {
		return true, "今日も一日お仕事お疲れさま♡毎日頑張って偉いね！"
	}
	m, _ = regexp.MatchString("^おはよう", tw.FullText)
	if m {
		return true, "おはようっ！今日も応援してるからね！"
	}
	return false, ""
}

type DevProc struct{}

func (p DevProc) Match(tw anaconda.Tweet) (bool, string) {
	m, _ := regexp.MatchString("さっそく", tw.FullText)
	if m {
		return m, "やるぞ " + time.Now().Format(timeStampFormat)
	}
	return m, ""
}

func (app *App) validateConf() bool {
	if app.SleepSeconds < 0 {
		return false
	}
	return true
}
