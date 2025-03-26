/*
history:
2020/3/27 v1

GoFmt
GoBuildNull

httpe api.cloudflare.com put /client/v4/zones/XXX/dns_records/YYY X-Auth-Email:$CloudflareEmail X-Auth-Key:$CloudflareKey Content-Type:application/json <<EOF
{
"type":"TXT",
"name":"_dnslink.iriy.de",
"content":"dnslink=/ipfs/QmZZZ",
"ttl":1
}
EOF
*/

package main

func main() {

	return
}
