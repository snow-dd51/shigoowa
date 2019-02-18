package main

import (
	"encoding/json"
	"fmt"
	"github.com/ChimeraCoder/anaconda"
	"io/ioutil"
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

type AppConf struct {
	AccessToken    string `json:"access-token"`
	AccessSecret   string `json:"access-secret"`
	ConsumerKey    string `json:"consumer-key"`
	ConsumerSecret string `json:"consumer-secret"`
}

func NewApp(confpath string) (*App, error) {
	content, err := ioutil.ReadFile(confpath)
	if err != nil {
		return nil, err
	}
	conf := AppConf{}
	err = json.Unmarshal(content, &conf)
	if err != nil {
		return nil, err
	}
	api := anaconda.NewTwitterApiWithCredentials(conf.AccessToken, conf.AccessSecret, conf.ConsumerKey, conf.ConsumerSecret)
	return &App{
		IsDebug:      true,
		SleepSeconds: 3,
		TwAPI:        api,
	}, nil
}

func (app *App) mainLoop() {
	if !app.validateConf() {
		return
	}
	ln := 0
	for true {
		fmt.Printf("[%s] Loop %d\n", time.Now().Format(timeStampFormat), ln)
		ln++
		time.Sleep(time.Duration(app.SleepSeconds) * time.Second)
	}
}

func (app *App) validateConf() bool {
	if app.SleepSeconds < 0 {
		return false
	}
	return true
}
