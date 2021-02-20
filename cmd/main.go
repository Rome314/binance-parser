package main

import (
	"binanceParser/cmd/article"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)


var (
	botToken = os.Getenv("BOT_TOKEN")
	chatId = os.Getenv("CHAT_ID")
)
func getText(a article.Article) string {

	title := strings.ReplaceAll(a.Title, "/", "\\/")
	title = strings.ReplaceAll(title, "(", "\\(")
	title = strings.ReplaceAll(title, ")", "\\)")
	title = strings.ReplaceAll(title, "-", "\\-")
	title = strings.ReplaceAll(title, "&", "\\&")
	title = strings.ReplaceAll(title, "!", "\\!")

	link := fmt.Sprintf("https://www.binance.com/en/support/announcement/%s", a.Code)
	text := fmt.Sprintf(`🔥 Новость на бинансе\!  [%s](%s)`, title, link)

	return text
}

func SendMessage(a article.Article) (err error) {
	text := getText(a)

	u := fmt.Sprintf(`https://api.telegram.org/%s/sendMessage?chat_id=%s&text=%s&parse_mode=MarkdownV2`, botToken, chatId, url.QueryEscape(text))

	resp, err := http.Get(u)
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusBadRequest {
		bts, _ := ioutil.ReadAll(resp.Body)
		log.Println(string(bts))
		err = fmt.Errorf(resp.Status)
		return err
	}
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf(resp.Status)
		return err
	}
	time.Sleep(time.Second)
	return nil

}

func main() {

	t := time.NewTicker(time.Second * 60)

	l := article.NewListener(t)

	finsi := make(chan int)
	a := make(chan article.Article)

	go func() {
		for {
			select {
			case art := <-a:
				err := SendMessage(art)
				if err != nil {
					log.Printf("[%d] err: %v", art.Id, err)
				} else {
					log.Printf("[%d] Success", art.Id)
				}
				break
			case _ = <-finsi:
				return

			}
		}

	}()

	err := l.Listen(finsi, a)
	if err != nil {
		log.Println(err)
		return
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	os.Exit(1)

}
