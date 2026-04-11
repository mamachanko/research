# Web client for Pi coding agent on a VPS (accessed locally via port-forwarding)

## Clarified objective
Run the **Pi coding agent** on a VPS, but use it from your local browser by forwarding a local port to the VPS.

## Recommended topology

```text
Local browser (http://localhost:9000)
   -> SSH local forward (-L 9000:127.0.0.1:9000)
      -> VPS loopback web client (127.0.0.1:9000)
         -> VPS loopback Pi agent API/runner (127.0.0.1:7777)
```

Design principles:
- Keep agent and UI private on VPS loopback only.
- Use SSH local forwarding for secure single-user access.
- Add reverse proxy + SSO only if you need multi-user/public hostname access.

## Build plan

### 1) Run Pi agent runtime behind a local API on VPS
Expose an internal API (REST/WebSocket) that launches and streams agent jobs. Bind to `127.0.0.1:7777` only.

### 2) Run a web UI on VPS
Serve the web app on `127.0.0.1:9000`. UI sends requests to `http://127.0.0.1:7777` (or `/api` proxied there).

### 3) Access from your laptop via SSH port-forward
From local machine:

```bash
ssh -N -L 9000:127.0.0.1:9000 user@your-vps.example
```

Then open `http://localhost:9000` locally.

### 4) Keep it persistent (optional)
Use systemd on VPS for agent/UI and (if needed) autossh from local jump host.

## Minimal reference implementation

### Docker Compose (loopback only)
- `pi-agent` listens on `127.0.0.1:7777`
- `pi-web` listens on `127.0.0.1:9000`

See: `examples/docker-compose.yml`

### Optional Caddy for hostname/TLS
If you need `pi.example.com`, terminate TLS at Caddy and proxy to loopback services.

See: `examples/Caddyfile`

### systemd unit for the Pi agent
Keeps the agent runner alive and restartable.

See: `examples/pi-agent.service`

## Security baseline
- Do not bind agent API to `0.0.0.0`.
- Use SSH keys (disable password auth if possible).
- Restrict VPS firewall to SSH and optional 443 only.
- Keep model/provider secrets in env files with strict permissions.
- Add per-session workspace isolation and execution limits in the agent runtime.

## If you need collaboration
Two patterns:
1. **SSH-forward only (recommended default):** safest and simplest.
2. **Public HTTPS app:** Caddy/Nginx + auth layer (OIDC) + rate limits + audit logs.

## Common gotchas
- White page after forward: UI bound to different VPS port than forwarded target.
- API errors from browser: CORS/proxy mismatch between UI and agent API.
- Works locally on VPS but not from laptop: SSH forward not active or blocked by local VPN policy.
- Random disconnects: use `ServerAliveInterval` options in SSH config.

## Quickstart checklist
1. Start services on VPS (`docker compose up -d`).
2. Verify on VPS: `curl http://127.0.0.1:9000/health` and `curl http://127.0.0.1:7777/health`.
3. Start local forward: `ssh -N -L 9000:127.0.0.1:9000 user@your-vps.example`.
4. Open `http://localhost:9000`.

## Bottom line
For Pi (the coding agent), the clean setup is: **run everything private on the VPS loopback and access it through local SSH `-L` forwarding**. It gives you a web UX locally without exposing the agent runtime directly to the internet.
