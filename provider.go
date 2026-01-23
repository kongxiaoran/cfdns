package main

import (
	"fmt"
)

type DNSProvider interface {
	GetRecordID(dnsName string, hostType string) (string, error)
	UpdateRecord(recordID string, dnsName string, hostType string, forwardName string) error
	AddRecord(dnsName string, hostType string, forwardName string) error
}

type ProviderType string

const (
	ProviderCloudflare ProviderType = "cloudflare"
	ProviderAliyun     ProviderType = "aliyun"
)

type DNSProviderFactory struct{}

func NewDNSProviderFactory() *DNSProviderFactory {
	return &DNSProviderFactory{}
}

func (f *DNSProviderFactory) CreateProvider(providerType ProviderType, config map[string]string) (DNSProvider, error) {
	switch providerType {
	case ProviderCloudflare:
		email, ok1 := config["email"]
		key, ok2 := config["key"]
		zoneID, ok3 := config["zoneID"]
		if !ok1 || !ok2 || !ok3 {
			return nil, fmt.Errorf("Cloudflare 配置缺少必要参数: email, key, zoneID")
		}
		return NewCloudflareProvider(email, key, zoneID), nil

	case ProviderAliyun:
		accessKeyId, ok1 := config["accessKeyId"]
		accessKeySecret, ok2 := config["accessKeySecret"]
		if !ok1 || !ok2 {
			return nil, fmt.Errorf("阿里云配置缺少必要参数: accessKeyId, accessKeySecret")
		}
		return NewAliyunProvider(accessKeyId, accessKeySecret)

	default:
		return nil, fmt.Errorf("不支持的 DNS 提供商类型: %s", providerType)
	}
}
