# ip-json

一个基于多个免费提供商的ip查询api


## 用法

```bash
# 查指定 IP
curl localhost:8080/8.8.8.8

# 查自己的 IP（部署到公网上才能用，本地跑会是127.0.0.1然后报错）
curl localhost:8080
```

返回类似：

```json
{
  "ip": "8.8.8.8",
  "hostname": "dns.google",
  "city": "Mountain View",
  "region": "California",
  "region_code": "CA",
  "country": "US",
  "country_name": "United States",
  "continent": "North America",
  "postal": "94043",
  "latitude": 37.4056,
  "longitude": -122.0775,
  "timezone": "America/Los_Angeles",
  "utc_offset": "-0800",
  "org": "Google LLC",
  "asn": "AS15169",
  "isp": "Google LLC",
  "is_proxy": false,
  "_sources": 5
}
```

### 错误情况

```bash
curl localhost:8080/abc          # 400不合法
curl localhost:8080/192.168.1.1  # 422是内网ip
```

## 数据源

| # | 数据源 | 免费额度 |
|---|--------|----------|
| 1 | ip-api.com | 45次/分钟 |
| 2 | ipapi.co | ~1000次/天 |
| 3 | ipwhois.app | 不限 |
| 4 | freeipapi.com | 不限 |
| 5 | ip2location.io | 500次/天 |
| 6 | db-ip.com | 不限 |

## 环境变量

都有默认值

| 变量 | 默认值 | 干嘛的 |
|------|--------|--------|
| `PORT` | `8080` | 监听端口 |
| `CACHE_TTL` | `10m` | 缓存多久过期 |
| `CACHE_MAX_SIZE` | `10000` | 最多缓存多少条 |
| `PROVIDER_TIMEOUT` | `3s` | 查上游的总超时 |
