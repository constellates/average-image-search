package main

import (
	"fmt"
	"net/http"
	"net/url"
	"io/ioutil"
	"encoding/json"
)

type Trend struct {
	Title               string `json:"title"`
	TitleLinkUrl        string `json:"titleLinkUrl"`
}

type Trends struct {
	Date          string `json:"date"`
	formattedDate string `json:"formattedDate"`
	TrendsList    []Trend `json:"trendsList"`
	Link        string `json:"link"`
}

type TrendResponse struct {
	SummaryMessage    string   `json:"summaryMessage"`
	// DataUpdateTime    float    `json:"dataUpdateTime"`
	TrendsByDateList  []Trends `json:"trendsByDateList"`
	OldestVisibleDate string   `json:"oldestVisibleDate"`
	LastPage          bool     `json:"lastPage"`
}

func getTrend(dateString string) (string) {
  fmt.Println("Getting trend for date ", dateString)

	// curl --data "ajax=1&pn=p1&htd=20170630&htv=l" https://trends.google.com/trends/hottrends/hotItems
	resp, err := http.PostForm("https://trends.google.com/trends/hottrends/hotItems", url.Values{"ajax": {"1"}, "pn": {"p1"}, "htd": {dateString}, "htv": {"l"}})
	if err != nil {
		panic(err.Error())
	}

	// extract body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err.Error())
	}

	// unmarshall json
	var s = new(TrendResponse)
	err = json.Unmarshal(body, &s)
	if err != nil {
		fmt.Println("whoops:", err)
	}

	fmt.Println("Yesterday's top trend is ", s.TrendsByDateList[0].TrendsList[0].Title)

	// return trend
	return s.TrendsByDateList[0].TrendsList[0].Title
}
