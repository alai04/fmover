package main

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	log "github.com/sirupsen/logrus"
)

var (
	k    = koanf.New(".")
	done = make(chan bool)
	wg   sync.WaitGroup
)

func main() {
	cfgFilename := "config.yml"
	f := file.Provider(cfgFilename)
	if err := k.Load(f, yaml.Parser()); err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	var tasks []taskStruct
	if err := k.Unmarshal("tasks", &tasks); err != nil {
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
