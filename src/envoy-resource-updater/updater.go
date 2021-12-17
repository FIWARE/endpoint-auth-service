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

	// Folder to write the envoy config to
	proxyConfigFolder = os.Getenv("PROXY_CONFIG_FOLDER")
	// Folder where the configmap, containing listener.yaml, cluster.yaml and envoy.yaml, is mounted
	configMapFolder = os.Getenv("CONFIG_MAP_FOLDER")
	// Should the container run as an init-container, e.g. also copy the envoy.yaml?
	runAsInit, err := strconv.ParseBool(os.Getenv("RUN_AS_INIT"))

	if err != nil {
		log.Fatal("Not able to read env var RUN_AS_INIT", err)
	}

	if proxyConfigFolder == "" {
		log.Fatal("No config folder was provided.")
	}

	if configMapFolder == "" {
		log.Fatal("No config map folder was provided.")
	}

	log.Info("Start updating: " + configMapFolder + " to " + proxyConfigFolder)

	if runAsInit {

		log.Info("Run as init-container.")
		updateStaticResources()
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
				log.Info("event:", event)
				if event.Op&fsnotify.Create == fsnotify.Create || event.Op&fsnotify.Write == fsnotify.Write {
					updateDynamicResources()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Info("error:", err)
			}
		}
	}()
	if err != nil {
		log.Fatal(err)
	}
	<-done
}

/**
* Update the static envoy.yaml from the config-map
 */
func updateStaticResources() {
	log.Info("Copy envoy.yaml")

	envoyYamlFile, err := ioutil.ReadFile(configMapFolder + "/envoy.yaml")
	if err != nil {
		log.Warn("Was not able to read envoy.yaml ", err)
	}
	err = os.WriteFile(proxyConfigFolder+"/envoy.yaml", envoyYamlFile, 0644)
	if err != nil {
		log.Warn("Was not able to write envoy.yaml ", err)
	}
}

// looks a little weird, but envoy listens to mv-events in the filesystem. To trigger such event,
// we copy the updated configmap to the config dir and than move(os.Rename) it to the configuration location.
func updateDynamicResources() {

	listenerYamlFile, err := ioutil.ReadFile(configMapFolder + "/listener.yaml")
	if err != nil {
		log.Warn("Was not able to read listener.yaml ", err)
		return
	}

	clusterYamlFile, err := ioutil.ReadFile(configMapFolder + "/cluster.yaml")
	if err != nil {
		log.Warn("Was not able to read cluster.yaml ", err)
		return
	}

	err = os.WriteFile(proxyConfigFolder+"/cluster.yaml.o", clusterYamlFile, 0644)
	if err != nil {
		log.Warn("Was not able to write cluster.yaml ", err)
		return
	}

	err = os.WriteFile(proxyConfigFolder+"/listener.yaml.o", listenerYamlFile, 0644)
	if err != nil {
		log.Warn("Was not able to write listener.yaml ", err)
		return
	}

	// First move the cluster yaml to trigger its reload, before the listeners are loaded.
	// If the order is the other way around, envoy can fail because the listener tries to use a cluster, that is
	// not available, yet.
	err = os.Rename(proxyConfigFolder+"/cluster.yaml.o", proxyConfigFolder+"/cluster.yaml")
	if err != nil {
		log.Warn("Was not able to move cluster.yaml.", err)
		return
	}
	err = os.Rename(proxyConfigFolder+"/listener.yaml.o", proxyConfigFolder+"/listener.yaml")
	if err != nil {
		log.Warn("Was not able to move listener.yaml.", err)
		return
	}

}
