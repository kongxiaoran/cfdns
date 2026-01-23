package main

import (
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

var dnsFactory *DNSProviderFactory

func acceptRequest(writer http.ResponseWriter, request *http.Request) {
	query := request.URL.Query()
	var dnsName = query.Get("dnsName")
	var forwardName = query.Get("forwardName")
	var idString = query.Get("id")
	var hostType = query.Get("hostType")
	id, err := strconv.Atoi(idString)
	if err != nil {
		http.Error(writer, "无效的 ID", http.StatusBadRequest)
		return
	}

	err = update(id, dnsName, hostType, forwardName)
	if err != nil {
		http.Error(writer, fmt.Sprintf("更新失败: %v", err), http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.Header().Set("Access-Control-Allow-Origin", "*")
	writer.WriteHeader(http.StatusOK)

	responseData := map[string]string{"message": "Data processed successfully"}
	json.NewEncoder(writer).Encode(responseData)
}

func getProvider(nodeIndex int) (DNSProvider, error) {
	if nodeIndex < 0 || nodeIndex >= len(DbJSON.NodeCollection) {
		return nil, fmt.Errorf("节点索引超出范围: %d", nodeIndex)
	}

	node := DbJSON.NodeCollection[nodeIndex]

	providerTypeStr := node.Provider
	if providerTypeStr == "" {
		providerTypeStr = string(ProviderCloudflare)
	}
	providerType := ProviderType(providerTypeStr)

	var config map[string]string
	if DbJSON.ProviderConfigs != nil {
		config = DbJSON.ProviderConfigs[providerTypeStr]
	}

	if config == nil {
		return nil, fmt.Errorf("未找到提供商配置: %s，请在 providerConfigs 中配置", providerTypeStr)
	}

	provider, err := dnsFactory.CreateProvider(providerType, config)
	if err != nil {
		return nil, fmt.Errorf("创建 DNS 提供商失败: %v", err)
	}

	return provider, nil
}

func update(id int, dnsName string, hostType string, forwardName string) error {
	provider, err := getProvider(id)
	if err != nil {
		return err
	}

	node := DbJSON.NodeCollection[id]
	providerTypeStr := node.Provider
	if providerTypeStr == "" {
		providerTypeStr = string(ProviderCloudflare)
	}
	providerType := ProviderType(providerTypeStr)

	recordID, err := provider.GetRecordID(dnsName, hostType)
	if err != nil {
		if providerType == ProviderAliyun || providerType == ProviderCloudflare {
			err = provider.AddRecord(dnsName, hostType, forwardName)
			if err != nil {
				fmt.Println(time.Now().Format("2006-01-02 15:04:05") + "  新增 DNS 记录失败: " + err.Error())
				return fmt.Errorf("新增 DNS 记录失败: %v", err)
			}
			fmt.Println(time.Now().Format("2006-01-02 15:04:05") + "  新增 " + dnsName + " 记录为：" + forwardName + "  成功! ")
			DbJSON.NodeCollection[id].ForwardName = forwardName + "#" + hostType
			UpdateToFile()
			return nil
		}
		fmt.Println(time.Now().Format("2006-01-02 15:04:05") + "  获取 DNS 记录 ID 失败: " + err.Error())
		return fmt.Errorf("获取 DNS 记录 ID 失败: %v", err)
	}

	err = provider.UpdateRecord(recordID, dnsName, hostType, forwardName)
	if err != nil {
		fmt.Println(time.Now().Format("2006-01-02 15:04:05") + "  更新 DNS 记录失败: " + err.Error())
		return fmt.Errorf("更新 DNS 记录失败: %v", err)
	}

	fmt.Println(time.Now().Format("2006-01-02 15:04:05") + "  将 " + dnsName + " 修改为：" + forwardName + "  成功! ")

	DbJSON.NodeCollection[id].ForwardName = forwardName + "#" + hostType
	UpdateToFile()

	return nil
}

func getPageDate(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	writer.Header().Set("Access-Control-Allow-Origin", "*")
	writer.WriteHeader(http.StatusOK)

	json.NewEncoder(writer).Encode(DbJSON)
}

//go:embed index.html
var content embed.FS

func main() {
	flag.StringVar(&ConfigPath, "config", "data.json", "配置文件的路径")
	flag.Parse()
	fmt.Println("配置文件路径:", ConfigPath)

	dnsFactory = NewDNSProviderFactory()

	GetDateFromFile()

	port := DbJSON.Port
	if port == "" {
		port = "8082"
	}
	fmt.Println("服务端口:", port)

	go func() {
		http.Handle("/", http.FileServer(http.FS(content)))
		http.HandleFunc("/api", acceptRequest)
		http.HandleFunc("/page-date", getPageDate)
		http.HandleFunc("/update-node", getPageDate)
		http.HandleFunc("/update-forward", getPageDate)

		err := http.ListenAndServe(":"+port, nil)
		if err != nil {
			log.Fatal("ListenAndServe: ", err)
		}
	}()

	fmt.Println("启动文件监听:", ConfigPath)
	go func() {
		StartFileWatcher()
	}()

	select {}
}
