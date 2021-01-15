package main

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	done = make(chan bool)
	wg   sync.WaitGroup
)

func main() {
	cfgFilename := "config.yml"
	viper.SetConfigFile(cfgFilename)
	viper.AddConfigPath(".")
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		log.Fatalf("error loading config: %v", err)
	}

	var tasks []taskStruct
	if err := viper.UnmarshalKey("tasks", &tasks); err != nil {
		log.Fatalf("error unmarshaling config: %v", err)
	}
	log.Infoln("tasks:", tasks)

	for i := range tasks {
		tasks[i].interval = time.Second
		tasks[i].Start()
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	select {
	case <-sigs:
		log.Infoln("Exiting...")
		close(done)
	}

	wg.Wait()
}
