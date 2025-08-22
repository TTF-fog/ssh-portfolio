package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

func GetAsyncData(url string, rc chan *http.Response, auth_token string) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", "Bearer "+auth_token)
	client := &http.Client{}
	data, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	rc <- data //i think thi puts the data into the channeL>?
}
func insertNth(s string, n int) string {
	var buffer bytes.Buffer
	var n_1 = n - 1
	var l_1 = len(s) - 1
	for i, rune := range s {
		buffer.WriteRune(rune)
		if i%n == n_1 && i != l_1 {
			buffer.WriteRune('\n')
		}
	}
	return buffer.String()
}
func loadFrameworks() []list.Item {
	items, _ := os.ReadDir("descs")
	var frameworks []list.Item
	for _, item := range items {
		name := item.Name()
		dat, _ := os.ReadFile(filepath.Join("descs", name))
		var framework Framework
		json.Unmarshal(dat, &framework)
		if framework.ExpandedDescriptionMDFile != "" {
			md, err := os.ReadFile(framework.ExpandedDescriptionMDFile)
			if err == nil {
				framework.ExpandedDescriptionMD = string(md)
			}
		}
		framework.progress = progress.New()
		frameworks = append(frameworks, &framework)

	}
	fmt.Println(frameworks)
	return frameworks
}
func incrementVisitsCounter(v_count *int) {
	dat, _ := os.ReadFile("visits.txt")
	var visits visits
	json.Unmarshal(dat, &visits)
	visits.Visits += 1
	*v_count = visits.Visits
	s, _ := json.Marshal(visits)
	os.WriteFile("visits.txt", s, 0644)
}

func cacheData(stats *UserStats, dailyStats *dailyUserStats) {
	hackChan := make(chan *http.Response)
	go GetAsyncData("https://hackatime.hackclub.com/api/v1/users/U093XFSQ344/stats", hackChan, authToken)
	response := <-hackChan
	go GetAsyncData("https://hackatime.hackclub.com/api/hackatime/v1/users/U093XFSQ344/statusbar/today", hackChan, authToken)
	responseDaily := <-hackChan
	if response.StatusCode == 200 {
		body, _ := io.ReadAll(response.Body)
		err := json.Unmarshal(body, &stats)
		var temp map[string]json.RawMessage
		err = json.Unmarshal(body, &temp)
		if err != nil {
			logger.log(fmt.Sprintf("error unmarshalling response: %s", err.Error()))
		}
		err = json.Unmarshal(temp["data"], &stats)
		if err != nil {
			logger.log(fmt.Sprintf("error unmarshalling response: %s", err.Error()))
		}
		if err != nil {
			panic(err)
		}
	} else {
		stats.HumanReadableTotal = "Failed To Get Data"
	}
	if responseDaily.StatusCode == 200 {
		body, _ := io.ReadAll(responseDaily.Body)
		err := json.Unmarshal(body, &dailyStats)
		var temp map[string]map[string]json.RawMessage
		json.Unmarshal(body, &temp)
		println(string(temp["data"]["grand_total"]))
		json.Unmarshal(temp["data"]["grand_total"], &dailyStats)
		if err != nil {
			panic(err)
		}
	} else {
		dailyStats.Text = "Failed to Retrieve!"
	}
	dailyStats.Text = dailyStats.Text + "\n Cached at " + time.Now().Format("3:04:05PM MST")
	logger.log("Cached hackatime data at" + time.Now().Format("2006-01-02 15:04:05"))

}

type fileLogger struct {
	file *os.File
	lock *sync.RWMutex
}

func (f *fileLogger) log(s string) {
	f.lock.Lock()
	defer f.lock.Unlock()
	fmt.Fprintln(f.file, s)
}
func loadBlogs() []list.Item {
	items, _ := os.ReadDir("blogs")
	var frameworks []list.Item
	for _, item := range items {
		name := item.Name()
		barl := strings.Split(name, "-")
		if len(barl) == 0 {
			logger.log("failed to load blog data at " + name)
			continue
		}
		logger.log(fmt.Sprintf("loading blog: %s", barl[1]))

		//unix_timestamp-name-description
		dat, err := os.ReadFile(filepath.Join("blogs", name))

		var article Article
		out, err := strconv.Atoi(barl[0])
		if err != nil {
			panic(err)
		}

		article.DatePublished = time.Unix(int64(out), 0)
		article.Name = barl[1]
		article.Desc = barl[2][0 : len(barl[2])-3]
		article.Body = string(dat)
		frameworks = append(frameworks, &article)
	}
	sort.Slice(frameworks, func(i, j int) bool {
		return frameworks[i].(*Article).DatePublished.Unix() > frameworks[j].(*Article).DatePublished.Unix()
	})

	return frameworks
}
