package controller

// DNS provider modules available for the ACME DNS-01 challenge
// (see the dnsProvider configmap option).
import (
	_ "github.com/caddy-dns/cloudflare"
	_ "github.com/caddy-dns/route53"
)
