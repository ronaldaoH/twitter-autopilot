package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/devices"
	"github.com/go-rod/rod/lib/input"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"time"
)

type Bot struct {
	Browser        *rod.Browser
	Page           *rod.Page
	Current_URL    string
	FollowersChan  chan []string
	Followers      [][]string
	Logger         chan string
	Influencer     string
	Cookies        bool
	CookiesSession []*proto.NetworkCookieParam
	User           User
	OutputCSV string
	DoubleAuth bool

}

type User struct {
	username string
	password string
	cookiesFile string
}

func (Bot *Bot) GoToURL_Wait(url string) {
	Bot.Page.Navigate(url)
	Bot.Page.WaitLoad()
}

func (Bot *Bot) SetCookies() {

	file_cookies, _ := os.Open(Bot.User.cookiesFile)
	cookies, _ := ioutil.ReadAll(file_cookies)

	pn := []*proto.NetworkCookieParam{}

	json.Unmarshal(cookies, &pn)

	Bot.Page.SetCookies(pn)
}

func (Bot *Bot) InitBrowser() {
	l := launcher.New()//.Delete("--headless")

	u, _ := l.Launch()

	Bot.Browser = rod.New().ControlURL(u) //.Trace(true)

	Bot.Browser.Connect()
	Bot.Logger <- "[+] Init Browser OK"

	//tweet , _ := page.Element(`[aria-label="Twittear"]`)

	//tweet.Click(proto.InputMouseButtonLeft)

}

func (Bot *Bot) LoginWOCookies() {
	_ = Bot.Page.Navigate("https://twitter.com/login")

	el, _ := Bot.Page.Element(`[name="session[username_or_email]"]`)
	el.Input(Bot.User.username)

	el, _ = Bot.Page.Element(`[type="password"]`)
	el.Input(Bot.User.password)

	el.Press(input.Enter)
	Bot.Page.WaitLoad()

	if(Bot.DoubleAuth == true){
		time.Sleep(time.Second * 20)
	}

	Bot.CreateCookies()


}

func (Bot *Bot) CreateCookies() {
	cookies, _ := Bot.Page.Cookies([]string{Bot.Current_URL})

	b, err := json.Marshal(cookies)
	if err != nil {
		fmt.Println(err)
		return
	}

	ioutil.WriteFile(Bot.User.cookiesFile, b, 0644)
}

func (Bot *Bot) Login() {
	if !Bot.Cookies {
		Bot.LoginWOCookies()
	} else {
		Bot.SetCookies()
	}
	Bot.Logger <- "[+] Cerramos login"

}

func (Bot *Bot) CreatePage() {
	current_url := proto.TargetCreateTarget{URL: Bot.Current_URL}
	Bot.Page, _ = Bot.Browser.Page(current_url)
	auxURL := fmt.Sprintf(`https://mobile.twitter.com/%s`, Bot.Influencer)

	device_ := devices.Device{
		Capabilities:   nil,
		UserAgent:      "Mozilla/5.0 (iPhone; CPU iPhone OS 13_7 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) CriOS/85.0.4183.92 Mobile/15E148 Safari/604.1",
		AcceptLanguage: "ES",
		Screen: devices.Screen{
			Horizontal: devices.ScreenSize{
				Width:  200,
				Height: 800,
			},
			Vertical: devices.ScreenSize{
				Width:  400,
				Height: 800,
			},
		},
	}
	Bot.Page.Emulate(device_)

	Bot.Page.Navigate(auxURL)
	Bot.Page.WaitLoad()
	Bot.Logger <- "[+] Cerramos CreatePage"
}

func (Bot *Bot) CreateCSV() {
	records := [][]string{[]string{"href", "name", "arroba"}}

	file, _ := os.Create(Bot.OutputCSV)
	defer file.Close()
	w := csv.NewWriter(file)

	for _, follower := range Bot.Followers {
		records = append(records, follower)
	}
	for _, record := range records {
		if err := w.Write(record); err != nil {
			log.Fatalln("error writing record to csv:", err)
		}
	}

	w.Flush()

	if err := w.Error(); err != nil {
		log.Fatal(err)
	}
	msg := fmt.Sprintf("[+] CSV READY! => %s" , Bot.OutputCSV)
	Bot.Logger <- msg

}
func (Bot *Bot) CloseChannels() {
	close(Bot.FollowersChan)
	close(Bot.Logger)
}

func (Bot *Bot) GetFollowers(num_records int) {
	auxURL := fmt.Sprintf(`https://mobile.twitter.com/%s/followers`, Bot.Influencer)
	Bot.Page.Navigate(auxURL)
	Bot.Page.WaitLoad()
	time.Sleep(time.Second * 3)

	var cont = 0
	for len(Bot.Followers) <= num_records {
		list, _ := Bot.Page.Elements(`[data-testid="primaryColumn"] a[role="link"]`)

		for _, item := range list {
			href := fmt.Sprintf("%s", item.MustProperty("href"))
			spans := item.MustElements("span")

			if len(spans) == 3 &&
				!contains(Bot.Followers, href) &&
				!strings.Contains(href, "%") &&
				!strings.Contains(href, ".co/") {
				name, errName := spans[0].Text()
				arroba, errArroba := spans[2].Text()
				if errName == nil && errArroba == nil {
					record := []string{href, name, arroba, Bot.Influencer}
					Bot.Logger <- fmt.Sprintf("Usuario Extraido => link : %s, Name :  %s, Arroba : %s", href, name, arroba)
					newSlice := make([]string, 3)
					copy(newSlice, record)
					if len(newSlice) != 0 {
						Bot.FollowersChan <- newSlice
						Bot.Followers = append(Bot.Followers, record)
						item.ScrollIntoView()
					}
					cont++
				}

			}
			if len(Bot.Followers) == num_records {
				return
			}
		}
		time.Sleep(time.Second * 1)
	}
	Bot.Logger <- "===== Terminó ====="
	//Bot.CloseChannels()
}

func main() {
	/*
	const numFollowers = 100

	bot := &Bot{}
	bot.Influencer = "chuisochuisez"
	bot.FollowersChan = make(chan []string, 1000)
	bot.Cookies = false
	bot.Logger = make(chan string)
	bot.OutputCSV = "./followers.csv"
	bot.User.cookiesFile = "./cookies.txt"
	bot.User.username = ""
	bot.User.password = ""
	bot.DoubleAuth = false

	go func() {
		bot.InitBrowser()
		bot.CreatePage()
		bot.Login()
		bot.GetFollowers(numFollowers)
		bot.CreateCSV()
		bot.CloseChannels()

	}()

	var msgClose = false
	//var followersClose = false
	var logger string
	var msg []string
	for {
		select {
		case msg, _ = <-bot.FollowersChan:
			if len(msg) != 0 {
				fmt.Println("[FOLLOWERS]", msg)
			}
		case logger, msgClose = <-bot.Logger:
			if logger != "" {
				fmt.Println("[LOGGER]", logger)
			}
		}
		if !msgClose {
			fmt.Println("bye bye!")
			break
		}

	}
	fmt.Println(len(bot.Followers))

*/


}

func contains(s [][]string, e string) bool {
	for _, a := range s {
		if a[0] == e {
			return true
		}
	}
	return false
}
