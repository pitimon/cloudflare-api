# cloudflare-dns-manager
---
- [Cloudflare User API Token](https://dash.cloudflare.com/profile/api-tokens)
---

>> BASH
```
cp cloudflare_credentials.org cloudflare_credentials
```
===
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