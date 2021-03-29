# twitter-autopilot

extract the followers of your favorite Twitter influencer.

```
const numFollowers = 100

bot := &Bot{}
bot.Influencer = "chuisochuisez"
bot.FollowersChan = make(chan []string, 1000)
bot.Cookies = false
bot.Logger = make(chan string)
bot.OutputCSV = "./chuisofollowers.csv"
bot.User.cookiesFile = "./chuiso.txt"
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

```
