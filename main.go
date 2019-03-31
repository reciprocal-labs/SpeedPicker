package main

import (
	"flag"
	"fmt"
	"speedPicker/board"
	"speedPicker/config"
	"speedPicker/httpserver"
	"time"
)

func main() {
	configFile := flag.String("config", "config.json", "configuration file path")
	flag.Parse()

	boardConf, err := config.Load(*configFile)
	if err != nil {
		panic(fmt.Sprintf("Error reading configuration file %s: %s", *configFile, err))
	}

	lockPanel := board.New(boardConf)
	go lockPanel.Run()

	go httpserver.Serve(lockPanel, boardConf.HttpAddr)

	for {
		time.Sleep(60 * time.Second)
	}

}
