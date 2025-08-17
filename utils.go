package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"net/http"
	"os"
	"path/filepath"
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
