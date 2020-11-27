package main

import (
	"github.com/spf13/viper"
	"log"
	"net/http"
	"strings"
	"time"
)

func main() {
	loadConfig()
	server := serverState{
		isChilling: true,
	}

	server.AddDeathAction(func() {
		server.isChilling = true
		log.Println("Server has died")
		restartServer()
	})

	server.AddDeathAction(func() {
		webhookerHeartbeatLost(server)
	})


	http.HandleFunc("/heartbeat", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received Heartbeat")

		server.lastHeartbeat = time.Now()
		server.isChilling = false
	})

	http.HandleFunc("/chill", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received a chill request")
		server.isChilling = true
	})

	go func() {
		checkInterval := viper.GetDuration("check-interval")
		for {
			time.Sleep(checkInterval)
			if !server.isChilling && server.IsDead() {
				server.RunDeathActions()
			}
		}
	}()

	addr := viper.GetString("address")
	log.Println("Listening on ", addr)
	http.ListenAndServe(addr, nil)
}

type serverState struct {
	lastHeartbeat time.Time
	isChilling    bool
	deathActions  []func()
}

func (s *serverState) IsDead() bool {
	return time.Since(s.lastHeartbeat) > viper.GetDuration("acceptable-heartbeat-delay")
}

func (s *serverState) RunDeathActions() {
	for _, action := range s.deathActions {
		action()
	}
}

func (s *serverState) AddDeathAction(action func()) {
	s.deathActions = append(s.deathActions, action)
}

func loadConfig() {
	viper.SetConfigName("heartbeat")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	viper.SetDefault("acceptable-heartbeat-delay", time.Second*10)
	viper.SetDefault("address", ":8080")
	viper.SetDefault("check-interval", time.Second*10)

	err := viper.ReadInConfig()

	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Println(err)
		}
	}
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
}


func restartServer() {
	log.Println("restarting server")

	req, _ := http.NewRequest("POST", viper.GetString("nanny-restart-url"), nil)
	req.Header.Add("Authorization", viper.GetString("nanny-auth"))
	client := http.Client{}

	_, err := client.Do(req)
	if err != nil {
		log.Println("server restart errored", err)
	}

	log.Println("restarted server")
}


func webhookerHeartbeatLost(s serverState) {
	url := viper.GetString("webhooker-url") + "/heartbeat-lost"

	log.Println("posting to webhooker", url)

	data, _ := json.Marshal(map[string]string{
		"realm": viper.GetString("server-name"),
		"last_heartbeat": s.lastHeartbeat.Unix()
	})

	buf = bytes.NewBuffer(data)

	resp, err := http.Post(url, "application/json", buf)
	if err != nil {
		log.Println("webhooker request failed ", err)
		return
	}
	defer resp.Body.Close()
}