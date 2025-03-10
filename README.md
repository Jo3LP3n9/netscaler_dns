# How to use this tool.
## Pre-requisitions
```
- Golang
- dependency package
- Netscaler credential
- Netscaler Nitro Enabled (w/ HTTP)
```
## Key Files
```
Current Path
 |---main.go
 |---nstoken.txt
```
# CLI command
```bash
go run main.go -f <token file> -rrtype <a|aaaa|mx|ns|soa|txt|srv|cname> -action <ADD|GET|DELETE> [arguments]
```
# Arguments list

## 參數說明:
```
  -rrtype：記錄類型（a, aaaa, ns, cname, srv, soa, txt）
  -action：操作類型（ADD, DELETE, GET）
  -f：Token檔案（預設為當前路徑下的 nstoken.txt）
```
## 各記錄類型所需參數:

### A 記錄:
```
  -hostname：主機名
  -ipaddress：IP地址
  -ttl：TTL
```
### 範例：
```bash
 go run main.go -rrtype a -action ADD -hostname "example.com" -ipaddress "192.0.2.1" -ttl 300
```
### AAAA 記錄:
```bash
  -hostname：主機名
  -ipv6address：IPv6地址
  -ttl：TTL
```
###  範例：
```bash
go run main.go -rrtype aaaa -action ADD -hostname "example.com" -ipv6address "2001:db8::1" -ttl 300
```
### CNAME 記錄:
```bash
  -hostname：別名
  -canonicalname：正規名稱
  -ttl：TTL
```
###  範例：
```bash
go run main.go -rrtype cname -action ADD -hostname "alias.example.com" -canonicalname "canonical.example.com" -ttl 300
```
### MX 記錄:
```bash
  -domain：域名
  -mx：MX記錄
  -pref：優先級
  -ttl：TTL
```
###  範例：
```bash
go run main.go -rrtype mx -action ADD -domain "example.com" -mx "mail.example.com" -pref 10 -ttl 300
```
### NS 記錄:
```bash
  -domain：域名
  -nameserver：名稱伺服器
  -ttl：TTL
```
###  範例：
```bash
go run main.go -rrtype ns -action ADD -domain "example.com" -nameserver "ns1.example.com" -ttl 300
```
### SOA 記錄:
```bash
  -domain：域名
  -originserver：原始伺服器
  -contact：聯絡人
  -serial：序列號
  -refresh：刷新時間
  -retry：重試時間
  -expire：過期時間
  -minimum：最小TTL
  -ttl：TTL
```
###  範例：
```bash
go run main.go -rrtype soa -action ADD -domain "example.com" -originserver "ns1.example.com" -contact "admin.example.com" -serial 2023032001 -refresh 3600 -retry 600 -expire 1209600 -minimum 300 -ttl 300
```
### SRV 記錄:
```bash
  -domain：域名
  -target：目標
  -priority：優先級
  -weight：權重
  -port：端口
  -ttl：TTL
```
###  範例：
```bash
go run main.go -rrtype srv -action ADD -domain "_service._tcp.example.com" -target "target.example.com" -priority 10 -weight 5 -port 80 -ttl 300
```
### TXT 記錄:
```bash
  -domain：域名
  -txtString：TXT記錄內容
  -ttl：TTL
```
###  範例：
```bash
go run main.go -rrtype txt -action ADD -domain "example.com" -txtString "v=spf1 include:_spf.example.com ~all" -ttl 300
```
