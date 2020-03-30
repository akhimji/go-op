/*
Copyright 2016 The Kubernetes Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Note: the example only works with the code within the same release/branch.
package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/fsnotify/fsnotify"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	//
	// Uncomment to load all auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth"
	//
	// Or uncomment to load specific auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/openstack"
)

const CONFIG_FILE = "/etc/config/configmap-microservice-demo.yaml"

type Config struct {
	Message string `yaml:"message"`
}

func watchfile(sendch chan<- string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()
	err = watcher.Add("/etc/config/configmap-microservice-demo.yaml")
	if err != nil {
		log.Fatal(err)
	}
	for {
		sendch <- "Ping"
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			log.Println("event:", event)
			sendch <- event.Name
			if event.Op&fsnotify.Write == fsnotify.Write {
				//log.Println("modified file:", event.Name)

			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("error:", err)
		default:
			sendch <- "End"
		}
	}
}

func listPods() {
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
	pods, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))
	_, err = clientset.CoreV1().Pods("istio-system").Get("istiod-7976d98b5-m6mkv", metav1.GetOptions{})
	if errors.IsNotFound(err) {
		fmt.Printf("Pod istiod-7976d98b5-m6mkv not found in istio-system namespace\n")
	} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
		fmt.Printf("Error getting pod %v\n", statusError.ErrStatus.Message)
	} else if err != nil {
		panic(err.Error())
	} else {
		fmt.Printf("Found istiod-7976d98b5-m6mkv pod in istio-system namespace\n")
	}
}
func main() {

	chnl := make(chan string)
	go watchfile(chnl)
	var msg string
	for {
		time.Sleep(4 * time.Second)
		msg = <-chnl
		if msg == "End" {
			listPods()
		} else if msg == "REMOVE" {
			fmt.Println("COnfig Change Restarting!")
			os.Exit(1)
		}
	}
}
