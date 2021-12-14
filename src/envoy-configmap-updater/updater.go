package main

import (
	"context"
	"io/ioutil"
	"os"

	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var proxyConfigFolder string
var configMap string
var configMapNamespace string

func main() {

	proxyConfigFolder = os.Getenv("PROXY_CONFIG_FOLDER")
	configMap = os.Getenv("PROXY_CONFIG_MAP")
	configMapNamespace = os.Getenv("PROXY_CONFIG_MAP_NAMESPACE")

	if proxyConfigFolder == "" {
		log.Fatal("No config folder was provided.")
	}

	if configMap == "" {
		log.Fatal("No config map was provided.")
	}

	if configMapNamespace == "" {
		log.Fatal("No config map namespace was provided.")
	}

	log.Info("Start watching " + proxyConfigFolder + " and push to " + configMapNamespace + "/" + configMap)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	err = watcher.Add(proxyConfigFolder)
	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				log.Println("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					updateConfigMap()
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

func updateConfigMap() {
	log.Warn("Update the map")

	listenerYamlFile, err := ioutil.ReadFile(proxyConfigFolder + "/listener.yaml")
	if err != nil {
		log.Printf("listenerYamlFile. Get err   #%v ", err)
	}

	clusterYamlFile, err := ioutil.ReadFile(proxyConfigFolder + "/cluster.yaml")
	if err != nil {
		log.Printf("clusterYamlFile. Get err   #%v ", err)
	}

	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	maps := clientset.CoreV1().ConfigMaps(configMapNamespace)
	// get the old map
	cm, err := maps.Get(context.TODO(), configMap, metav1.GetOptions{})
	if err != nil {
		log.Warn("No map", err)
	}

	cm.Data["listener.yaml"] = string(listenerYamlFile)
	cm.Data["cluster.yaml"] = string(clusterYamlFile)

	cm, err = maps.Update(context.TODO(), cm, metav1.UpdateOptions{})
	if err != nil {
		log.Warn("Was not able to update map", err)
	}
	log.Warn(cm.Data)

}
