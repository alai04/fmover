package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
)

var k = koanf.New(".")

type taskStruct struct {
	From         string `koanf:"from"`
	To           string `koanf:"to"`
	interval     time.Duration
	lastFileInfo []os.FileInfo
}

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
	fmt.Printf("tasks: %v\n", tasks)

	for i := range tasks {
		tasks[i].interval = time.Second
		tasks[i].Start()
	}
	<-make(chan bool)
}

func (t *taskStruct) Start() (err error) {
	go t.Run(context.Background())

	log.Println("watching", t.From)
	return nil
}

func (t *taskStruct) Run(ctx context.Context) {
	for {
		fn, err := t.checkFromDir()
		if err == nil && fn != "" {
			log.Println("Moving file", fn, "to dest dir...")
			if err = t.copyFile(fn); err != nil {
				log.Println("Error copying file ", fn, ":", err)
			} else {
				if err = os.Remove(filepath.Join(t.From, fn)); err != nil {
					log.Println("Error removing file ", fn, ":", err)
				}
			}
		} else {
			time.Sleep(t.interval)
		}
	}
}

func (t *taskStruct) checkFromDir() (filename string, err error) {
	var curFileInfo []os.FileInfo
	for {
		log.Println("Checking directory", t.From)
		curFileInfo, err = ioutil.ReadDir(t.From)
		if err != nil {
			log.Println("Error checking", t.From, ":", err)
			return
		}
		filename = findSameSize(t.lastFileInfo, curFileInfo)
		if len(filename) > 0 {
			return
		}
		t.lastFileInfo = curFileInfo
		time.Sleep(t.interval)
	}
}

func (t *taskStruct) copyFile(filename string) (err error) {
	src, err := os.Open(filepath.Join(t.From, filename))
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := ioutil.TempFile(t.To, ".tmp_")
	if err != nil {
		return err
	}
	dstName := dst.Name()
	defer os.Remove(dstName)
	_, err = io.Copy(dst, src)
	dst.Close()
	if err != nil {
		return err
	}
	return os.Rename(dstName, filepath.Join(t.To, filename))
}

func findSameSize(lastFileInfo, curFileInfo []os.FileInfo) (filename string) {
	for i, j := 0, 0; i < len(lastFileInfo) && j < len(curFileInfo); {
		name1, name2 := lastFileInfo[i].Name(), curFileInfo[j].Name()
		if name1 > name2 {
			j++
		} else if name2 > name1 {
			i++
		} else {
			if !lastFileInfo[i].IsDir() && !curFileInfo[j].IsDir() &&
				lastFileInfo[i].Size() == curFileInfo[j].Size() && !strings.HasPrefix(name1, ".") {
				return name1
			}
			i++
			j++
		}
	}
	return
}
