package main

import (
	"time"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	log "github.com/sirupsen/logrus"
)

var k = koanf.New(".")

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
	<-make(chan bool)
}
