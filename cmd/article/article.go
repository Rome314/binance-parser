package article

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

const hostname = "https://www.binance.com"
var firebaseToken =os.Getenv("FIREBASE_TOKEN")
var firebaseUrl = "https://binance-parser-c76c9-default-rtdb.europe-west1.firebasedatabase.app/rest/binance.json?auth="+firebaseToken

type articleListener struct {
	lastId int32
	t      *time.Ticker
	l      func(t *time.Ticker, finish chan int, articles chan Article)
}

func NewListener(t *time.Ticker) *articleListener {

	lId := getLastId()
	log.Printf("LAST ID: %d", lId)
	return &articleListener{lastId: lId, t: t}
}
func getLastId() int32 {
	res, err := http.Get(firebaseUrl)
	if err != nil {
		log.Printf("GET LAST ID ERR %v", err)
	}

	bts, _ := ioutil.ReadAll(res.Body)

	s := map[string]int32{}

	_ = json.Unmarshal(bts, &s)
	return s["last_id"]

}
func (a *articleListener) setLastId(id int32) {

	i := map[string]int32{
		"last_id": id,
	}

	bts, _ := json.Marshal(i)

	body := bytes.NewReader(bts)
	req, _ := http.NewRequest(http.MethodPut, firebaseUrl, body)
	_, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("SET LAST ID ERR %v", err)
	}
}

func (a *articleListener) Listen(finish chan int, articles chan Article) (err error) {

	a.l = a.listener

	go a.l(a.t, finish, articles)
	return nil

}

func (a *articleListener) listener(t *time.Ticker, finish chan int, articles chan Article) {
	for {
		select {
		case _ = <-t.C:
			list, err := a.getList()
			if err != nil {
				err = fmt.Errorf("get list: %v", err)
				log.Println(err)
				break
			}

			if a.lastId == 0 {
				for i := len(list) - 1; i >= 0; i-- {
					article := list[i]
					a.lastId = article.Id
					a.setLastId(article.Id)
					articles <- article

				}
				break
			}

			start := false
			for i := len(list) - 1; i >= 0; i-- {
				article := list[i]
				if article.Id == a.lastId {
					start = true
					continue
				}
				if start {
					log.Printf("NEW: %d", article.Id)
					a.lastId = article.Id
					a.setLastId(article.Id)
					articles <- article

				}
			}
			break
		case _ = <-finish:
			return

		}
	}

}

func (a *articleListener) getList() (articles []Article, err error) {
	articles = []Article{}
	u := fmt.Sprintf("%s/gateway-api/v1/public/cms/article/catalog/list/query?catalogId=48&pageNo=1&pageSize=15", hostname)

	resp, err := http.Get(u)
	if err != nil {
		err = fmt.Errorf("API: %v", err)
		return
	}

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf(resp.Status)
		return
	}

	if resp.Body == nil {
		err = fmt.Errorf("nil body")
		return
	}

	respStruct := getResponse{}

	defer resp.Body.Close()
	bts, _ := ioutil.ReadAll(resp.Body)

	err = json.Unmarshal(bts, &respStruct)
	if err != nil {
		err = fmt.Errorf("unmarshaling: %v", err)
		return
	}

	return respStruct.Data.Articles, nil

}
