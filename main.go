package main

import (
	"fmt"
	"github.com/ChimeraCoder/anaconda"
	"github.com/snow-dd51/shigoowa/conf"
	"net/url"
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
}

func Debugf(msg string, arg ...interface{}) {
	a := make([]interface{}, 1)
	a[0] = time.Now().Format(timeStampFormat)
	a = append(a, arg)
	fmt.Printf("[%s] "+msg+"\n", a...)
}

func NewApp(confpath string) (*App, error) {
	ac := conf.NewAppConf()
	err := ac.Read(confPath)
	if err != nil {
		return nil, err
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
		IsDebug:      true,
		SleepSeconds: 61,
		TwAPI:        api,
	}, nil
}

func (app *App) mainLoop() {
	if !app.validateConf() {
		return
	}
	ln := 0
	// 再起動に備えて開始地点は保存したい
	lastStatusId := ""
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
		}
		if len(tl) > 0 {
			lastStatusId = tl[0].IdStr

			Debugf("%v", lastStatusId)
		}
		// ここで処理する
		if lastStatusId == "" {
			break
		}
		time.Sleep(time.Duration(app.SleepSeconds) * time.Second)
	}
}

func (app *App) validateConf() bool {
	if app.SleepSeconds < 0 {
		return false
	}
	return true
}
