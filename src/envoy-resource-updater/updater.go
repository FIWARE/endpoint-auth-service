package main

import (
	"io/ioutil"
	"os"
	"strconv"

	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
)

var proxyConfigFolder string
var configMapFolder string

func main() {

	proxyConfigFolder = os.Getenv("PROXY_CONFIG_FOLDER")
	configMapFolder = os.Getenv("CONFIG_MAP_FOLDER")
	runAsInit, err := strconv.ParseBool(os.Getenv("RUN_AS_INIT"))

	if err != nil {
		log.Fatal("Not able to read env var RUN_AS_INIT", err)
	}

	if runAsInit {
		updateDynamicResources()
		return
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	err = watcher.Add(configMapFolder)
	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				log.Println("event:", event)
				if event.Op&fsnotify.Create == fsnotify.Create {
					updateDynamicResources()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()
	if err != nil {
		log.Fatal(err)
	}
	<-done
}

// logs a little weird, but envoy listens to mv-events in the filesystem. To trigger such event,
// we copy the updated configmap to the config dir and than move(os.Rename) it to the configuration location.
func updateDynamicResources() {
	log.Println("Update dynamic resources")

	listenerYamlFile, err := ioutil.ReadFile(configMapFolder + "/listener.yaml")
	if err != nil {
		log.Printf("listenerYamlFile. Get err   #%v ", err)
	}

	clusterYamlFile, err := ioutil.ReadFile(configMapFolder + "/cluster.yaml")
	if err != nil {
		log.Printf("clusterYamlFile. Get err   #%v ", err)
	}

	envoyYamlFile, err := ioutil.ReadFile(configMapFolder + "/envoy.yaml")
	if err != nil {
		log.Printf("clusterYamlFile. Get err   #%v ", err)
	}

	err = os.WriteFile(proxyConfigFolder+"/listener.yaml.o", listenerYamlFile, 0644)
	if err != nil {
		log.Warn("Was not able to copy listener.yaml.", err)
	}

	err = os.WriteFile(proxyConfigFolder+"/cluster.yaml.o", clusterYamlFile, 0644)
	if err != nil {
		log.Warn("Was not able to copy cluster.yaml.", err)
	}

	err = os.WriteFile(proxyConfigFolder+"/envoy.yaml.o", envoyYamlFile, 0644)
	if err != nil {
		log.Warn("Was not able to copy envoy.yaml.", err)
	}

	err = os.Rename(proxyConfigFolder+"/listener.yaml.o", proxyConfigFolder+"/listener.yaml")
	if err != nil {
		log.Warn("Was not able to move listener.yaml.", err)
	}
	err = os.Rename(proxyConfigFolder+"/cluster.yaml.o", proxyConfigFolder+"/cluster.yaml")
	if err != nil {
		log.Warn("Was not able to move cluster.yaml.", err)
	}
	err = os.Rename(proxyConfigFolder+"/envoy.yaml.o", proxyConfigFolder+"/envoy.yaml")
	if err != nil {
		log.Warn("Was not able to move envoy.yaml.", err)
	}
}
