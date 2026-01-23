package main

import (
	"fmt"
	"strings"

	alidns20150109 "github.com/alibabacloud-go/alidns-20150109/v5/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	credential "github.com/aliyun/credentials-go/credentials"
)

type AliyunProvider struct {
	client *alidns20150109.Client
}

func NewAliyunProvider(accessKeyId, accessKeySecret string) (*AliyunProvider, error) {
	credentialConfig := &credential.Config{
		Type:            tea.String("access_key"),
		AccessKeyId:     tea.String(accessKeyId),
		AccessKeySecret: tea.String(accessKeySecret),
	}

	cred, err := credential.NewCredential(credentialConfig)
	if err != nil {
		return nil, fmt.Errorf("创建阿里云凭据失败: %v", err)
	}

	config := &openapi.Config{
		Credential: cred,
	}
	config.Endpoint = tea.String("alidns.aliyuncs.com")

	client, err := alidns20150109.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("创建阿里云客户端失败: %v", err)
	}

	return &AliyunProvider{
		client: client,
	}, nil
}

func (p *AliyunProvider) GetRecordID(dnsName string, hostType string) (string, error) {
	domainName, rr, err := parseDomainName(dnsName)
	if err != nil {
		return "", err
	}

	describeDomainRecordsRequest := &alidns20150109.DescribeDomainRecordsRequest{
		DomainName: tea.String(domainName),
		RRKeyWord:  tea.String(rr),
		Type:       tea.String(hostType),
	}

	runtime := &util.RuntimeOptions{}

	resp, err := p.client.DescribeDomainRecordsWithOptions(describeDomainRecordsRequest, runtime)
	if err != nil {
		return "", fmt.Errorf("查询 DNS 记录失败: %v", err)
	}

	if resp.Body == nil || resp.Body.DomainRecords == nil {
		return "", fmt.Errorf("响应数据为空")
	}

	records := resp.Body.DomainRecords.Record
	if records == nil || len(records) == 0 {
		return "", fmt.Errorf("未找到 DNS 记录: %s (类型: %s)", dnsName, hostType)
	}

	recordID := records[0].RecordId
	if recordID == nil {
		return "", fmt.Errorf("记录 ID 为空")
	}

	return tea.StringValue(recordID), nil
}

func (p *AliyunProvider) UpdateRecord(recordID string, dnsName string, hostType string, forwardName string) error {
	_, rr, err := parseDomainName(dnsName)
	if err != nil {
		return err
	}

	updateDomainRecordRequest := &alidns20150109.UpdateDomainRecordRequest{
		RecordId: tea.String(recordID),
		RR:       tea.String(rr),
		Type:     tea.String(hostType),
		Value:    tea.String(forwardName),
	}

	runtime := &util.RuntimeOptions{}

	_, err = p.client.UpdateDomainRecordWithOptions(updateDomainRecordRequest, runtime)
	if err != nil {
		return fmt.Errorf("更新 DNS 记录失败: %v", err)
	}

	return nil
}

func (p *AliyunProvider) AddRecord(dnsName string, hostType string, forwardName string) error {
	domainName, rr, err := parseDomainName(dnsName)
	if err != nil {
		return err
	}

	addDomainRecordRequest := &alidns20150109.AddDomainRecordRequest{
		Lang:       tea.String("zh"),
		DomainName: tea.String(domainName),
		Type:       tea.String(hostType),
		RR:         tea.String(rr),
		Value:      tea.String(forwardName),
	}

	runtime := &util.RuntimeOptions{}

	_, err = p.client.AddDomainRecordWithOptions(addDomainRecordRequest, runtime)
	if err != nil {
		return fmt.Errorf("新增 DNS 记录失败: %v", err)
	}

	return nil
}

func parseDomainName(fullDomainName string) (domainName string, rr string, err error) {
	parts := strings.Split(fullDomainName, ".")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("域名格式不正确: %s", fullDomainName)
	}

	if len(parts) == 2 {
		return fullDomainName, "", nil
	}

	domainName = strings.Join(parts[len(parts)-2:], ".")
	rr = strings.Join(parts[:len(parts)-2], ".")

	return domainName, rr, nil
}
