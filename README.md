# cloudflare-ddns

This was a project designed to explore the [Cloudflare API](https://api.cloudflare.com/)
through their official [Go Client](https://github.com/cloudflare/cloudflare-go).

This application updates a given domain name to the current machine's ip address
with the use of the https://ipify.org service. Alternative ip address services
may be used if it returns the ip address of the current machine in plain text
by overriding the configuration option.

You can run this service in a cron job or a systemd timer, or once, up to you!

## Installation

Install via with the Go toolchain to compile from source:

```
go get github.com/wyattjoh/cloudflare-ddns
```

Download pre-compiled binary on the [Releases Page](https://github.com/wyattjoh/cloudflare-ddns/releases/latest) for your Arch/OS.

## Configuration

Configuration is specified in the environment or as command line arguments.

- `-key` or `ENV['CF_API_KEY']` (*required*) - specify the Global (not CA) Cloudflare API Key generated on the ["My Account" page](https://www.cloudflare.com/a/account/my-account).
- `-email` or `ENV['CF_API_EMAIL']` (*required*) - Email address associated with your Cloudflare account.
- `-domain` or `ENV['CF_DOMAIN']` (*required*) - Domain name in question that you want to update. (i.e. `mypage.example.com` OR `example.com`)
- `-ipendpoint` or `ENV['CF_IP_ENDPOINT']` (optional, default: `https://api.ipify.org`) - Alternative ip address service endpoint.
