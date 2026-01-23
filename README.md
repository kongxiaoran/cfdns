# CFDNS - DNS 记录管理工具

CFDNS 是一个基于 Go 语言开发的 DNS 记录管理工具，支持动态调整节点域名的 DNS 记录。目前支持 Cloudflare 和阿里云 DNS 服务，并提供简单的 Web 界面进行管理。

## 功能特性

- 支持多 DNS 提供商（Cloudflare、阿里云）
- Web 界面管理，操作简单直观
- 配置文件热重载，修改后自动生效
- 支持 CNAME、A、AAAA 等记录类型
- 阿里云支持自动创建不存在的记录

## 项目结构

```
cfdns/
├── src/
│   ├── main.go                # 程序入口和 HTTP 服务
│   ├── provider.go            # DNS 提供商接口定义
│   ├── cloudflare_provider.go # Cloudflare 实现
│   ├── aliyun_provider.go     # 阿里云实现
│   ├── db.go                  # 配置文件读写
│   └── index.html             # Web 前端页面
├── data.json.example          # 配置文件示例
├── data.json                 # 配置文件（需自行创建）
├── go.mod                    # Go 模块依赖
├── go.sum                    # Go 模块校验
└── README.md                 # 项目说明
```

## 安装

### 从 GitHub Releases 下载（推荐）

访问 [GitHub Releases](https://github.com/kongxiaoran/cfdns/releases) 页面，下载对应平台的二进制文件。

#### Linux/macOS

```bash
# 下载最新版本（请替换为实际版本号）
wget https://github.com/kongxiaoran/cfdns/releases/latest/download/cfdns_Linux_x86_64.tar.gz

# 解压
tar -xzf cfdns_Linux_x86_64.tar.gz

# 移动到系统路径
sudo mv cfdns /usr/local/bin/

# 赋予执行权限
sudo chmod +x /usr/local/bin/cfdns
```

#### Windows

下载 `cfdns_Windows_x86_64.zip`，解压后将 `cfdns.exe` 添加到系统 PATH 或直接使用。

### 从源码编译

#### 环境要求

- Go 1.21 或更高版本

#### 克隆项目

```bash
git clone https://github.com/kongxiaoran/cfdns.git
cd cfdns
```

#### 下载依赖

```bash
go mod download
```

#### 编译

Linux 平台：

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o cfdns
```

macOS 平台：

```bash
go build -o cfdns
```

Windows 平台：

```bash
GOOS=windows GOARCH=amd64 go build -o cfdns.exe
```

## 配置

1. 复制示例配置文件：

```bash
cp data.json.example data.json
```

2. 编辑 `data.json`，填入实际配置信息：

```json
{
  "nodeCollection": [
    {
      "name": "节点1",
      "dnsName": "node1.example.com",
      "forwardName": "target.example.com#CNAME",
      "provider": "cloudflare"
    }
  ],
  "forwardCollection": [
    {
      "name": "转发目标1",
      "forwardName": "forward1.example.com",
      "hostType": "CNAME"
    }
  ],
  "providerConfigs": {
    "cloudflare": {
      "email": "your-email@example.com",
      "key": "your-cloudflare-api-key",
      "zoneID": "your-zone-id"
    },
    "aliyun": {
      "accessKeyId": "your-access-key-id",
      "accessKeySecret": "your-access-key-secret"
    }
  }
}
```

### 配置说明

| 字段 | 说明 |
|------|------|
| `nodeCollection` | 需要管理的 DNS 节点列表 |
| `forwardCollection` | 可用的转发目标列表 |
| `providerConfigs` | DNS 提供商的 API 配置 |

#### 节点配置 (nodeCollection)

| 字段 | 说明 | 示例 |
|------|------|------|
| `name` | 节点显示名称 | "新加坡节点" |
| `dnsName` | 完整域名 | "sg.example.com" |
| `forwardName` | 当前转发目标 | "target.example.com#CNAME" |
| `provider` | DNS 提供商 | "cloudflare" 或 "aliyun" |

#### 转发配置 (forwardCollection)

| 字段 | 说明 | 示例 |
|------|------|------|
| `name` | 显示名称 | "优选IP-电信" |
| `forwardName` | 目标地址 | "1.2.3.4" |
| `hostType` | 记录类型 | "A" 或 "CNAME" |

#### Cloudflare 配置

在 Cloudflare 控制台获取：
- `email`: 账户邮箱
- `key`: API Key（My Profile -> API Tokens -> Global API Key）。官方页面地址：https://dash.cloudflare.com/profile/api-tokens
- `zoneID`: 域名 Zone ID（Domains -> 选择域名 -> 右侧显示）

#### 阿里云配置

在阿里云访问控制 RAM 获取：
- `accessKeyId`: AccessKey ID
- `accessKeySecret`: AccessKey Secret

## 使用

### Docker 方式（推荐）

```bash
# 拉取镜像
docker pull ghcr.io/kongxiaoran/cfdns:latest

# 运行容器（需要先创建配置文件）
docker run -d \
  --name cfdns \
  -p 8082:8082 \
  -v $(pwd)/data.json:/etc/cfdns/data.json \
  ghcr.io/kongxiaoran/cfdns:latest
```

### 二进制文件方式

#### 启动服务

前台运行：

```bash
./cfdns -config data.json
```

后台运行：

```bash
nohup ./cfdns -config data.json > cfdns.log 2>&1 &
```

#### 访问 Web 界面

默认监听端口：`8082`

访问地址：`http://localhost:8082`

在 Web 界面中选择节点后，从下拉菜单选择目标转发地址即可完成 DNS 更新。

### 命令行参数

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `-config` | 配置文件路径 | data.json |

### API 接口

| 接口 | 方法 | 说明 |
|------|------|------|
| `/` | GET | Web 前端页面 |
| `/api` | GET | 更新 DNS 记录 |
| `/page-date` | GET | 获取页面数据 |
| `/update-node` | GET | 更新节点（预留） |
| `/update-forward` | GET | 更新转发（预留） |

### 检查运行状态

```bash
netstat -tunlp | grep 8082
```

## 安全建议

1. 不要将 `data.json` 提交到版本控制系统
2. 限制 API Key 的权限，仅赋予必要的 DNS 管理权限
3. 生产环境建议使用 HTTPS
4. 定期更换 API 密钥

## 许可证

MIT License
