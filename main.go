package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type configs struct {
	APIKey              string `json:"api_key"`
	ZoneID              string `json:"zone_identifier"`
	ID                  string `json:"id"`
	IntervalS           int    `json:"interval_secs"`
	TwitchRedisCacheURL string `json:"twitch_redis_cache_url"`
}

type actionValue struct {
	URL        string `json:"url"`
	StatusCode int    `json:"status_code"`
}
type action struct {
	ID    string      `json:"id"`
	Value actionValue `json:"value"`
}
type pageRule struct {
	Actions []action `json:"actions"`
}

func main() {
	config := getConfig()
	checkLiveChannelsAndUpdate(config)

	intervalS, err := time.ParseDuration(fmt.Sprintf("%ds", config.IntervalS))
	if err != nil {
		log.Fatalln(err)
	}
	ticker := time.NewTicker(intervalS)
	quit := make(chan struct{})

	for {
		select {
		case <-ticker.C:
			checkLiveChannelsAndUpdate(config)
		case <-quit:
			ticker.Stop()
			return
		}
	}
}

func checkLiveChannelsAndUpdate(config configs) {
	twitchUserNames := getTwitchUserNames()
	liveList := getOnline(config, twitchUserNames)
	updateLink(config, liveList)
}

func getConfig() configs {
	jsonFile, err := os.Open("config.json")
	if err != nil {
		log.Fatalln(err)
	}

	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)
	var configValue configs
	json.Unmarshal(byteValue, &configValue)
	return configValue
}

func getTwitchUserNames() []string {
	type userNames []string

	jsonFile, err := os.Open("usernames.json")
	if err != nil {
		log.Fatalln(err)
	}

	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)
	var twitchUserNames userNames
	json.Unmarshal(byteValue, &twitchUserNames)
	return twitchUserNames
}

func getOnline(config configs, twitchUserNames []string) []string {
	type response struct {
		IsLive bool `json:"isLive"`
	}
	var liveList []string
	for _, username := range twitchUserNames {
		url := fmt.Sprintf("%s/%s", config.TwitchRedisCacheURL, username)
		resp, err := http.Get(url)
		if err != nil {
			log.Fatalln(err)
		}

		defer resp.Body.Close()

		var r response
		err = json.NewDecoder(resp.Body).Decode(&r)
		if err != nil {
			log.Fatalln(err)
		}

		if r.IsLive == true {
			liveList = append(liveList, username)
		}
	}
	if len(liveList) < 1 {
		liveList = twitchUserNames
	}
	log.Println(liveList)
	return liveList
}

func updateLink(config configs, liveList []string) {
	liveListStr := strings.Join(liveList, "/")
	liveListURL := fmt.Sprintf("http://multitwitch.tv/%s", liveListStr)

	putBody := pageRule{
		Actions: []action{
			{
				ID: "forwarding_url",
				Value: actionValue{
					URL:        liveListURL,
					StatusCode: 302,
				},
			},
		},
	}
	b, err := json.Marshal(putBody)

	client := &http.Client{}
	URL := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/pagerules/%s", config.ZoneID, config.ID)
	req, err := http.NewRequest(http.MethodPut, URL, bytes.NewBuffer(b))
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.APIKey))
	if err != nil {
		log.Fatalln(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println(string(body))
}
