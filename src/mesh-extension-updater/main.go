package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
)

var meshYamlFolder string
var meshYamlFile string

func main() {

	// Folder to read cluster and listener.yaml from
	meshYamlFolder = os.Getenv("MESH_CONFIG_FOLDER")
	meshYamlFile = os.Getenv("MESH_EXTENSION_FILE_NAME")

	if meshYamlFolder == "" {
		log.Fatal("No mesh-yaml folder was provided.")
		return
	}

	if meshYamlFile == "" {
		log.Fatal("No mesh-yaml file was provided.")
		return
	}

	log.Infof("Start watching %s.", meshYamlFolder)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("Was not able to create a new folder watcher. %v", err)
		return
	}

	err = watcher.Add(meshYamlFolder)
	if err != nil {
		log.Fatalf("Was not able to add watcher for folder %s. %v", meshYamlFolder, err)
		return
	}
	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				log.Println("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
					patchMeshExtension()
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

func patchMeshExtension() {

	ctx := context.TODO()
	// creates the in-cluster config
	cfg, err := rest.InClusterConfig()
	if err != nil {
		log.Warnf("Was not able to create an in-cluster config. %v", err)
		return
	}

	// read extension
	b, _ := ioutil.ReadFile(meshYamlFolder + "/" + meshYamlFile)
	obj := &unstructured.Unstructured{}

	dec := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	_, gvk, err := dec.Decode(b, nil, obj)

	if err != nil {
		log.Warnf("Was not able to decode the yaml file. %v", err)
	}

	fmt.Println(obj.GetName(), gvk.String())

	// 1. Prepare a RESTMapper to find GVR
	dc, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		log.Warnf("Was not able to create the rest mapper. %v", err)
		return
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))

	// 2. Prepare the dynamic client
	dyn, err := dynamic.NewForConfig(cfg)
	if err != nil {
		log.Warnf("Was not able to create the dynamic client. %v", err)
		return
	}

	// 4. Find GVR
	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		log.Warnf("Was not able to find the gvr. %v", err)
		return
	}

	// 5. Obtain REST interface for the GVR
	var dr dynamic.ResourceInterface
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		// namespaced resources should specify the namespace
		dr = dyn.Resource(mapping.Resource).Namespace(obj.GetNamespace())
	} else {
		// for cluster-wide resources
		dr = dyn.Resource(mapping.Resource)
	}

	// 6. Marshal object into JSON
	data, err := json.Marshal(obj)
	if err != nil {
		log.Warnf("Was not able to marshal the object to json. %v", err)
		return
	}

	// 7. Create or Update the object with SSA
	//     types.ApplyPatchType indicates SSA.
	//     FieldManager specifies the field owner ID.
	_, err = dr.Patch(ctx, obj.GetName(), types.ApplyPatchType, data, metav1.PatchOptions{
		FieldManager: "mesh-extension-updater",
	})
	if err != nil {
		log.Warnf("Was not able to patch the service mesh extension. %v", err)
	}
}
