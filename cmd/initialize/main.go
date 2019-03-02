package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/ChimeraCoder/anaconda"
	"github.com/snow-dd51/shigoowa/conf"
	"os"
	"regexp"
)

var (
	confPathp = flag.String("conf", "./auth.json", "config file path")
)

func main() {
	flag.Parse()
	ac := conf.NewAppConf()
	err := ac.Read(*confPathp)
	if err != nil {
		fmt.Printf("the file %s is not present nor a valid json file.\ncreate a new config file.\n", *confPathp)
	}
	sc := bufio.NewScanner(os.Stdin)
	if ac.ConsumerKey == "" {
		fmt.Println("consumer-key is empty")
		fmt.Print("input consumer-key: ")
		ret := sc.Scan()
		if !ret {
			fmt.Println("input error")
			return
		}
		ck := string(sc.Bytes())
		if ck == "" {
			fmt.Println("consumer-key is empty")
			return
		}
		ac.ConsumerKey = ck
	} else {
		fmt.Printf("consumer-key is %s\n", ac.ConsumerKey)
	}
	if ac.ConsumerSecret == "" {
		fmt.Println("consumer-secret is empty")
		fmt.Print("input consumer-secret: ")
		ret := sc.Scan()
		if !ret {
			fmt.Println("input error")
			return
		}
		cs := string(sc.Bytes())
		if cs == "" {
			fmt.Println("consumer-secret is empty")
			return
		}
		ac.ConsumerSecret = (cs)
	} else {
		fmt.Println("consumer-secret is set")
	}
	// ここまできたら途中経過保存してもよくない？
	// deferされたときのacの内容については要確認
	defer ac.Write(*confPathp)
	api := anaconda.NewTwitterApiWithCredentials("", "", ac.ConsumerKey, ac.ConsumerSecret)
	url, tmpCred, err := api.AuthorizationURL("")
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	fmt.Printf("URL: %s\n", url)
	fmt.Print("Input PIN: ")
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
}
