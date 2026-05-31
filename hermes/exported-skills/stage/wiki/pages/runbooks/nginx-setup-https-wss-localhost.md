---
title: "Nginx Setup for HTTPS and WSS on Localhost"
type: "runbook"
slug: "runbooks/nginx-setup-https-wss-localhost"
freshness: "2024-08-25T23:32:00Z"
tags:
  - "localhost"
  - "nginx"
  - "self-signed-certificate"
  - "ssl"
  - "websocket"
owners: []
source_revision_ids:
  - "srcrev_9ce38fff10eed080be0c8f242d8bda2b"
conflict_state: "none"
---

# Nginx Setup for HTTPS and WSS on Localhost

## Summary

A step-by-step guide to configure nginx as a reverse proxy for HTTPS and WebSocket Secure (WSS) endpoints on localhost using self-signed certificates, with geth as the backend.

## Claims

- Install nginx using Homebrew: brew install nginx `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/nginx-setup-wss-https-localhost-22a53dd922bf42d7a53404c31cddac6b#chunk-1) `source_document_id=srcdoc_860cd74094ba7f190f32725aacf5be2e` `source_revision_id=srcrev_9ce38fff10eed080be0c8f242d8bda2b` `chunk_id=srcchunk_17e4b223a62e4d76e6042dbf734cde76` `native_locator=https://www.notion.so/nginx-setup-wss-https-localhost-22a53dd922bf42d7a53404c31cddac6b#chunk-1` `source_timestamp=2024-08-25T23:32:00Z`
- Create a Subject Alternative Name (SAN) configuration file at ~/san.conf with the following content: [req] distinguished_name = req_distinguished_name x509_extensions = v3_req prompt = no [req_distinguished_name] CN = localhost [v3_req] subjectAltName = @alt_names [alt_names] DNS.1 = localhost IP.1 = 127.0.0.1 `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/nginx-setup-wss-https-localhost-22a53dd922bf42d7a53404c31cddac6b#chunk-1) `source_document_id=srcdoc_860cd74094ba7f190f32725aacf5be2e` `source_revision_id=srcrev_9ce38fff10eed080be0c8f242d8bda2b` `chunk_id=srcchunk_17e4b223a62e4d76e6042dbf734cde76` `native_locator=https://www.notion.so/nginx-setup-wss-https-localhost-22a53dd922bf42d7a53404c31cddac6b#chunk-1` `source_timestamp=2024-08-25T23:32:00Z`
- SANs are required because browsers and tools like curl validate certificates against the domain/IP; the CN field is ignored nowadays. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/nginx-setup-wss-https-localhost-22a53dd922bf42d7a53404c31cddac6b#chunk-1) `source_document_id=srcdoc_860cd74094ba7f190f32725aacf5be2e` `source_revision_id=srcrev_9ce38fff10eed080be0c8f242d8bda2b` `chunk_id=srcchunk_17e4b223a62e4d76e6042dbf734cde76` `native_locator=https://www.notion.so/nginx-setup-wss-https-localhost-22a53dd922bf42d7a53404c31cddac6b#chunk-1` `source_timestamp=2024-08-25T23:32:00Z`
- Generate a self-signed certificate and key using OpenSSL with the command: openssl req -x509 -nodes -days 365 -newkey rsa:2048 -key /usr/local/etc/nginx/ssl/nginx.key -out /usr/local/etc/nginx/ssl/nginx.crt -config ~/san.conf `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/nginx-setup-wss-https-localhost-22a53dd922bf42d7a53404c31cddac6b#chunk-1) `source_document_id=srcdoc_860cd74094ba7f190f32725aacf5be2e` `source_revision_id=srcrev_9ce38fff10eed080be0c8f242d8bda2b` `chunk_id=srcchunk_17e4b223a62e4d76e6042dbf734cde76` `native_locator=https://www.notion.so/nginx-setup-wss-https-localhost-22a53dd922bf42d7a53404c31cddac6b#chunk-1` `source_timestamp=2024-08-25T23:32:00Z`
- The nginx configuration file (nginx.conf) should be modified to include a server block listening on port 443 ssl for both IPv4 and IPv6, with server_name localhost, and ssl_certificate and ssl_certificate_key pointing to the generated files. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/nginx-setup-wss-https-localhost-22a53dd922bf42d7a53404c31cddac6b#chunk-1) `source_document_id=srcdoc_860cd74094ba7f190f32725aacf5be2e` `source_revision_id=srcrev_9ce38fff10eed080be0c8f242d8bda2b` `chunk_id=srcchunk_17e4b223a62e4d76e6042dbf734cde76` `native_locator=https://www.notion.so/nginx-setup-wss-https-localhost-22a53dd922bf42d7a53404c31cddac6b#chunk-1` `source_timestamp=2024-08-25T23:32:00Z`
- The location / block proxies requests to http://127.0.0.1:8545 for non-websocket connections and to http://127.0.0.1:8546/ for websocket connections, with the Upgrade and Connection headers set. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/nginx-setup-wss-https-localhost-22a53dd922bf42d7a53404c31cddac6b#chunk-1) `source_document_id=srcdoc_860cd74094ba7f190f32725aacf5be2e` `source_revision_id=srcrev_9ce38fff10eed080be0c8f242d8bda2b` `chunk_id=srcchunk_17e4b223a62e4d76e6042dbf734cde76` `native_locator=https://www.notion.so/nginx-setup-wss-https-localhost-22a53dd922bf42d7a53404c31cddac6b#chunk-1` `source_timestamp=2024-08-25T23:32:00Z`
- It is critical to use a trailing slash in the proxy_pass directive for the websocket endpoint (http://127.0.0.1:8546/), otherwise it will not work correctly. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/nginx-setup-wss-https-localhost-22a53dd922bf42d7a53404c31cddac6b#chunk-1) `source_document_id=srcdoc_860cd74094ba7f190f32725aacf5be2e` `source_revision_id=srcrev_9ce38fff10eed080be0c8f242d8bda2b` `chunk_id=srcchunk_17e4b223a62e4d76e6042dbf734cde76` `native_locator=https://www.notion.so/nginx-setup-wss-https-localhost-22a53dd922bf42d7a53404c31cddac6b#chunk-1` `source_timestamp=2024-08-25T23:32:00Z`
- A separate location /site block proxies to http://127.0.0.1:8080/ for a local web server. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/nginx-setup-wss-https-localhost-22a53dd922bf42d7a53404c31cddac6b#chunk-1) `source_document_id=srcdoc_860cd74094ba7f190f32725aacf5be2e` `source_revision_id=srcrev_9ce38fff10eed080be0c8f242d8bda2b` `chunk_id=srcchunk_17e4b223a62e4d76e6042dbf734cde76` `native_locator=https://www.notion.so/nginx-setup-wss-https-localhost-22a53dd922bf42d7a53404c31cddac6b#chunk-1` `source_timestamp=2024-08-25T23:32:00Z`
- Run geth with websockets and HTTP enabled: ./geth --local --http --ws.origins "*" --ws --ws.addr "localhost" --ws.port 8546 `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/nginx-setup-wss-https-localhost-22a53dd922bf42d7a53404c31cddac6b#chunk-2) `source_document_id=srcdoc_860cd74094ba7f190f32725aacf5be2e` `source_revision_id=srcrev_9ce38fff10eed080be0c8f242d8bda2b` `chunk_id=srcchunk_decb519393e08f8cd283eab16230715e` `native_locator=https://www.notion.so/nginx-setup-wss-https-localhost-22a53dd922bf42d7a53404c31cddac6b#chunk-2` `source_timestamp=2024-08-25T23:32:00Z`
- After configuration, reload nginx with sudo nginx -s reload and verify with nginx -t. `claim:claim_1_10` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/nginx-setup-wss-https-localhost-22a53dd922bf42d7a53404c31cddac6b#chunk-2) `source_document_id=srcdoc_860cd74094ba7f190f32725aacf5be2e` `source_revision_id=srcrev_9ce38fff10eed080be0c8f242d8bda2b` `chunk_id=srcchunk_decb519393e08f8cd283eab16230715e` `native_locator=https://www.notion.so/nginx-setup-wss-https-localhost-22a53dd922bf42d7a53404c31cddac6b#chunk-2` `source_timestamp=2024-08-25T23:32:00Z`
- Test the setup by visiting https://localhost/site, running a JSON-RPC curl command to https://localhost, and connecting via wscat to wss://localhost with --no-check flag. `claim:claim_1_11` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/nginx-setup-wss-https-localhost-22a53dd922bf42d7a53404c31cddac6b#chunk-2) `source_document_id=srcdoc_860cd74094ba7f190f32725aacf5be2e` `source_revision_id=srcrev_9ce38fff10eed080be0c8f242d8bda2b` `chunk_id=srcchunk_decb519393e08f8cd283eab16230715e` `native_locator=https://www.notion.so/nginx-setup-wss-https-localhost-22a53dd922bf42d7a53404c31cddac6b#chunk-2` `source_timestamp=2024-08-25T23:32:00Z`
- The guide assumes a macOS environment, using Homebrew and paths like /usr/local/etc/nginx/. `claim:claim_1_12` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/nginx-setup-wss-https-localhost-22a53dd922bf42d7a53404c31cddac6b#chunk-1) `source_document_id=srcdoc_860cd74094ba7f190f32725aacf5be2e` `source_revision_id=srcrev_9ce38fff10eed080be0c8f242d8bda2b` `chunk_id=srcchunk_17e4b223a62e4d76e6042dbf734cde76` `native_locator=https://www.notion.so/nginx-setup-wss-https-localhost-22a53dd922bf42d7a53404c31cddac6b#chunk-1` `source_timestamp=2024-08-25T23:32:00Z`

## Sources

- `source_document_id`: `srcdoc_860cd74094ba7f190f32725aacf5be2e`
- `source_revision_id`: `srcrev_9ce38fff10eed080be0c8f242d8bda2b`
- `source_url`: [Notion source](https://www.notion.so/nginx-setup-wss-https-localhost-22a53dd922bf42d7a53404c31cddac6b)
