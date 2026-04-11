# Pi web client via VPS port-forwarding

## Goal
Design a web-based client that you can open locally (in your browser) while the app is hosted on a VPS and communicates with a Raspberry Pi behind NAT.

## Recommended architecture (simple + robust)
Use a **reverse tunnel from the Pi to the VPS** so the VPS can always reach Pi-local services.

```text
[Your Browser] --> HTTPS --> [VPS reverse proxy + web app] --> localhost:18xxx (tunnel endpoint on VPS) ==> SSH reverse tunnel ==> [Pi service localhost:3000]
```

Why this is practical:
- No inbound port opening on home router.
- Pi only needs outbound SSH to VPS.
- VPS exposes only HTTPS; tunnel stays private on loopback.

## Reference implementation

### 1) Run your web client on VPS
Example Node/React backend on `127.0.0.1:8080` and reverse proxy via Caddy/Nginx.

### 2) Create persistent reverse tunnel on Pi
Install autossh:

```bash
sudo apt update
sudo apt install -y autossh
```

Create dedicated key and copy to VPS tunnel user:

```bash
ssh-keygen -t ed25519 -f ~/.ssh/pi_vps_tunnel -N ''
ssh-copy-id -i ~/.ssh/pi_vps_tunnel.pub tunnel@your-vps.example
```

Run tunnel (Pi -> VPS):

```bash
autossh -M 0 -N \
  -o ServerAliveInterval=30 \
  -o ServerAliveCountMax=3 \
  -o ExitOnForwardFailure=yes \
  -i ~/.ssh/pi_vps_tunnel \
  -R 127.0.0.1:18080:127.0.0.1:3000 \
  tunnel@your-vps.example
```

This maps `VPS 127.0.0.1:18080` to `Pi 127.0.0.1:3000`.

### 3) Reverse proxy routes
Example Caddy snippet:

```caddy
pi.example.com {
  encode zstd gzip

  # web UI hosted on VPS
  handle_path / {
    reverse_proxy 127.0.0.1:8080
  }

  # API calls proxied to Pi through tunnel
  handle_path /pi-api/* {
    uri strip_prefix /pi-api
    reverse_proxy 127.0.0.1:18080
  }
}
```

### 4) Make tunnel resilient (systemd on Pi)

```ini
[Unit]
Description=Pi to VPS reverse SSH tunnel
After=network-online.target
Wants=network-online.target

[Service]
User=pi
ExecStart=/usr/bin/autossh -M 0 -N -o ServerAliveInterval=30 -o ServerAliveCountMax=3 -o ExitOnForwardFailure=yes -i /home/pi/.ssh/pi_vps_tunnel -R 127.0.0.1:18080:127.0.0.1:3000 tunnel@your-vps.example
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

## Security checklist
- Use a dedicated VPS user (`tunnel`) with `nologin` shell.
- Restrict authorized key in `~tunnel/.ssh/authorized_keys` with options like `permitopen="127.0.0.1:18080",no-agent-forwarding,no-pty`.
- Keep tunnel bind on `127.0.0.1` (not `0.0.0.0`).
- Put authentication in front of web UI (OIDC, basic auth, or access proxy).
- Enforce HTTPS and automatic cert renewals.
- Rate-limit and audit access logs on VPS.

## Alternative: WireGuard mesh
If you want private IP connectivity among local machine, VPS, and Pi:
- run WireGuard peers on all three,
- expose Pi service only on WG interface,
- have VPS web app call Pi over WG private IP.

Tradeoff: slightly more setup, cleaner network model for multi-service deployments.

## Common failure modes
- Tunnel up but app unreachable: wrong local Pi port in `-R` mapping.
- Tunnel flaps: missing keepalive settings, unstable DNS, or host key mismatch.
- 502 from reverse proxy: VPS endpoint points to wrong forwarded port.
- Works internally but not externally: TLS / DNS / firewall misconfiguration on VPS.

## Minimal rollout plan
1. Confirm Pi service works locally (`curl localhost:3000`).
2. Bring up reverse tunnel and verify on VPS (`curl 127.0.0.1:18080`).
3. Add `/pi-api` reverse-proxy route and test from VPS host.
4. Expose HTTPS hostname and test from local browser.
5. Add auth + firewall hardening + monitoring.

## Conclusion
The most straightforward way is: **Pi initiates outbound reverse SSH tunnel to VPS**, VPS hosts the web client, and proxy routes Pi API traffic through the localhost tunnel endpoint. This avoids home NAT headaches while keeping the Pi service non-public.
