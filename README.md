# cloudflare-dns-manager
---
- [Cloudflare User API Token](https://dash.cloudflare.com/profile/api-tokens)
---

>> Before use
```
cp cloudflare_credentials.org cloudflare_credentials
cp go/cloudflare_credentials-master.json go/cloudflare_credentials.json 
```
---
>> Go Language
```
cd go
```
>> Initial go environments
สร้างโมดูล Go:
```
go mod init cloudflare-dns-manager
```
ติดตั้งไลบรารี Cloudflare:
```
go get github.com/cloudflare/cloudflare-go
```