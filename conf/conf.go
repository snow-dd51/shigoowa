package conf

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type AppConf struct {
	AccessToken    string `json:"access-token"`
	AccessSecret   string `json:"access-secret"`
	ConsumerKey    string `json:"consumer-key"`
	ConsumerSecret string `json:"consumer-secret"`
}

func NewAppConf() *AppConf {
	c := AppConf{}
	return &c
}

func (c *AppConf) Write(path string) error {
	json, err := json.Marshal(c)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(path, json, os.ModePerm&0644)
	return err
}

func (c *AppConf) Read(path string) error {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	err = json.Unmarshal(content, c)
	if err != nil {
		return err
	}
	return nil
}
