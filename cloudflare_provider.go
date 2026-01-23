package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type CloudflareProvider struct {
	email  string
	key    string
	zoneID string
}

func NewCloudflareProvider(email, key, zoneID string) *CloudflareProvider {
	return &CloudflareProvider{
		email:  email,
		key:    key,
		zoneID: zoneID,
	}
}

type CloudflareDnsInfo struct {
	Result []struct {
		ID string `json:"id"`
	} `json:"result"`
}

type CloudflareDnsUpdateReq struct {
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
	TTL     int    `json:"ttl"`
	Proxied bool   `json:"proxied"`
}

func (p *CloudflareProvider) GetRecordID(dnsName string, hostType string) (string, error) {
	client := &http.Client{}
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records?type=%s&name=%s",
		p.zoneID, hostType, dnsName)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Add("X-Auth-Email", p.email)
	req.Header.Add("X-Auth-Key", p.key)
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %v", err)
	}

	var dnsInfo CloudflareDnsInfo
	err = json.Unmarshal(body, &dnsInfo)
	if err != nil {
		return "", fmt.Errorf("解析响应失败: %v", err)
	}

	if len(dnsInfo.Result) == 0 {
		return "", fmt.Errorf("未找到 DNS 记录: %s (类型: %s)", dnsName, hostType)
	}

	return dnsInfo.Result[0].ID, nil
}

func (p *CloudflareProvider) UpdateRecord(recordID string, dnsName string, hostType string, forwardName string) error {
	data := CloudflareDnsUpdateReq{
		Name:    dnsName,
		Type:    hostType,
		Content: forwardName,
		TTL:     120,
		Proxied: false,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("序列化数据失败: %v", err)
	}

	client := &http.Client{}
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records/%s",
		p.zoneID, recordID)

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Add("X-Auth-Email", p.email)
	req.Header.Add("X-Auth-Key", p.key)
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("更新失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (p *CloudflareProvider) AddRecord(dnsName string, hostType string, forwardName string) error {
	data := CloudflareDnsUpdateReq{
		Name:    dnsName,
		Type:    hostType,
		Content: forwardName,
		TTL:     120,
		Proxied: false,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("序列化数据失败: %v", err)
	}

	client := &http.Client{}
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records", p.zoneID)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Add("X-Auth-Email", p.email)
	req.Header.Add("X-Auth-Key", p.key)
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("新增记录失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	return nil
}
