# vault-redirector

Simple Go app to redirect [Hashicorp Vault](https://www.vaultproject.io/) requests to the active node in a HA cluster.

## Status

__WARNING:__ Work was begun on this, but it is not complete or even really usable. This was my first-ever foray into Go. While I was able to get the basic logic working, it's become apparent that getting a production-ready, fully tested, documented service isn't going to happen within my time constraints. As such, I'm working on implementing the same functionality in Python's Twisted networking framework. That interim version can be seen at [https://github.com/jantman/vault-redirector-twisted](https://github.com/jantman/vault-redirector-twisted); initial capacity tests show performance within 25% of the Go variant at 1000 requests per second, with a _lower_ maximum response time.

## Purpose

There's a bit of a gap in usability of [Vault](https://www.vaultproject.io/) in a [High Availability](https://www.vaultproject.io/docs/concepts/ha.html) mode, at least in AWS:

* Vault's HA architecture is based on an active/standby model; only one server can be active at a time, and any others are standby. Standby servers respond to all API requests with a 307 Temporary Redirect to the Active server, but can only do this if they're unsealed (in the end of the [HA docs](https://www.vaultproject.io/docs/internals/high-availability.html): "It is important to note that only unsealed servers act as a standby. If a server is still in the sealed state, then it cannot act as a standby as it would be unable to serve any requests should the active server fail.").
* HashiCorp recommends managing infrastructure individually, i.e. not in an auto-scaling group. In EC2, if you want to run Consul on the same nodes, this is an absolute requirement as Consul requires static IP addresses in order for disaster recovery to work without downtime and manual changes.

As a result, we're left with a conundrum:

1. We can't put Vault behind an Elastic Load Balancer, because that would cause all API requests to appear to have the ELB's source IP address. Not only does this render any of the IP- or subnet-based authorization methods unusable, but it also means we lose client IPs in the audit logs (which is likely a deal-breaker for anyone concerned with security).
2. The only remaining option for HA, at least in AWS, is to use Route53 round-robin DNS records that have the IPs of all of the cluster members. This poses a problem because if one node in an N-node cluster is either offline or sealed, approximately 1/N of all client requests will be directed to that node and fail.

While it would be good for all clients to automatically retry these requests, it appears that most client libraries (and even the ``vault`` command line client) do not currently support this. While retry logic would certainly be good to implement in any case, it adds latency to retrieving secrets (in the common case where the cluster is reachable, but some nodes are down) and also does not account for possible DNS caching issues. Furthermore, we're providing Vault as a service to our organization; relying on retries would mean either adding retry logic to every Vault client library and getting those changes merged, or deviating from our plan of "here's your credentials and endpoint, see the upstream docs for your language's client library."

The best solution to this problem would be for [Vault issue #799](https://github.com/hashicorp/vault/issues/799), a request to add [PROXY Protocol](http://www.haproxy.org/download/1.5/doc/proxy-protocol.txt) support to Vault, to be completed. Both [AWS ELBs](http://docs.aws.amazon.com/ElasticLoadBalancing/latest/DeveloperGuide/enable-proxy-protocol.html) and HAProxy support this, and it would alleviate issue #1 above, allowing us to run Vault behind a load balancer but still have access to the original client IP address.

This small service is intended to provide an interim workaround until that solution is implemented.

## Functionality

We take advantage of Vault's 307 redirects (and the assumption that any protocol-compliant client library will honor them). Instead of connecting directly to the Vault service, clients connect to a load-balanced daemon running on the Vault nodes. This daemon asynchronously polls Consul for the health status of the Vault instances, and therefore knows the currently-active Vault instance at all times. All incoming HTTP(S) requests are simply 307 redirected to the active instance. As this service can safely be load balanced, it will tolerate failed nodes better than round-robin DNS. Since it redirects the client to the active node, the client's IP address will be properly seen by Vault.

## Requirements

In order to determine the active Vault instance, ``vault-redirector`` requires that Consul be running and monitoring the health of all Vault instances. Redirection can be to either the IP address or Consul node name running the active service.

Here is example of the [Consul service definition](https://www.consul.io/docs/agent/services.html) that we use (note we're running Vault with TLS):

```json
{
  "service":{
    "name": "vault",
    "tags": ["secrets"],
    "port": 8200,
    "check": {
      "id": "api",
      "name": "HTTPS API check on port 8200",
      "http": "https://127.0.0.1:8200/v1/sys/health",
      "interval": "5s",
      "timeout" : "2s"
    }
  }
}
```

## Build

You'll need Go 1.4+ (1.6+ highly recommended) installed and setup per [http://golang.org/doc/code.html](http://golang.org/doc/code.html), and you'll need ``$GOPATH/bin`` in your ``PATH``. The build process is as similar to that of [Vault](https://www.vaultproject.io/docs/install/install.html) as possible; in fact we're using Vault's build tooling to ensure as close platform compatibility as possible.

1. Clone this git repository into your GOPATH: [https://github.com/manheim/vault-redirector.git](https://github.com/manheim/vault-redirector.git)
2. Verify that the file ``$GOPATH/src/github.com/manheim/vault-redirector/main.go`` exists. If it doesn't, then you didn't clone vault-redirector into the proper path.
3. Run ``make dev`` to generate the binary for your current system.

## Installation

Simply place the

## Usage

TODO: usage

## Logging and Debugging

TODO.

## Testing

TODO.

## License

vault-redirector is licensed under the MIT license; see ``LICENSE`` for the text of the license. The build tooling (``Makefile`` and ``scripts/``) are from the Vault project and licensed under the Mozilla Public License, version 2.0; see ``LICENSE-build-tools`` for the text of the license.
