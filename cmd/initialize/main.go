package main

import (
	"bufio"
	"fmt"
	"github.com/ChimeraCoder/anaconda"
	"github.com/snow-dd51/shigoowa/conf"
	"os"
	"regexp"
)

const (
	confPath = "./auth.json"
)

func main() {
	ac := conf.NewAppConf()
	err := ac.Read(confPath)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	if ac.ConsumerKey == "" {
		fmt.Println("consumer-key is empty")
		return
	}
	if ac.ConsumerSecret == "" {
		fmt.Println("consumer-secret is empty")
		return
	}
	api := anaconda.NewTwitterApiWithCredentials("", "", ac.ConsumerKey, ac.ConsumerSecret)
	url, tmpCred, err := api.AuthorizationURL("")
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	fmt.Printf("URL: %s\n", url)
	fmt.Print("Input PIN: ")
	sc := bufio.NewScanner(os.Stdin)
	ret := sc.Scan()
	if !ret {
		fmt.Println("scannning PIN error")
		return
	}
	bpin := sc.Bytes()
	match, err := regexp.Match("^[0-9]+$", bpin)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	pin := string(bpin)
	if !match {
		fmt.Printf("Invalid PIN: %s\n", pin)
		return
	}
	fmt.Printf("Valid PIN: %s\n", pin)
	cred, value, err := api.GetCredentials(tmpCred, pin)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	fmt.Printf("Auth : %s (%s)\n", value["screen_name"][0], value["user_id"][0])
	ac.AccessToken = cred.Token
	ac.AccessSecret = cred.Secret
	ac.Write(confPath)
}
