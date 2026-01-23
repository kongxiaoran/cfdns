package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/fsnotify/fsnotify"
)

type NodeDB struct {
	NodeCollection    []Node    `json:"nodeCollection"`
	ForwardCollection []Forward `json:"forwardCollection"`
	ProviderConfigs   map[string]map[string]string `json:"providerConfigs,omitempty"`
	Port             string    `json:"port,omitempty"`
}

type Node struct {
	Name        string `json:"name"`
	DNSName     string `json:"dnsName"`
	ForwardName string `json:"forwardName"`
	Provider    string `json:"provider,omitempty"`
}

type Forward struct {
	Name        string `json:"name"`
	ForwardName string `json:"forwardName"`
	HostType    string `json:"hostType"`
}

var DbJSON NodeDB
var ConfigPath string

func GetDateFromFile() {
	fmt.Println(os.Getwd())
	file, err := os.Open(ConfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("文件不存在")
		} else {
			fmt.Println("打开文件时出错:", err)
		}
		return
	}

	byteValue, _ := ioutil.ReadAll(file)
	json.Unmarshal(byteValue, &DbJSON)

	if DbJSON.ProviderConfigs == nil {
		DbJSON.ProviderConfigs = make(map[string]map[string]string)
	}
}

func UpdateToFile() {
	jsonData, err := json.MarshalIndent(DbJSON, "", "  ")
	if err != nil {
		fmt.Println("生成JSON时出错:", err)
		return
	}

	file, _ := os.Create(ConfigPath)
	file.Write(jsonData)
	fmt.Println("JSON数据已写入文件")
}

func StartFileWatcher() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					fmt.Println("文件被修改,重新加载配置文件:", event.Name)
					GetDateFromFile()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(ConfigPath)
	if err != nil {
		log.Fatal(err)
	}
	<-make(chan struct{})
}
