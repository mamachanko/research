# Pi coding agent on VPS only (local access via SSH port-forward)

## Requirement
Pi is **not installed on your local machine**. Pi runs **only on the VPS**.
Your local machine is just a client (browser + SSH).

## Architecture that matches this requirement

```text
Local laptop
  - Browser
  - SSH client
      |
      | ssh -N -L 9000:127.0.0.1:9000 user@vps
      v
VPS
  - Pi agent runtime/service (127.0.0.1:7777)
  - Web UI/API gateway (127.0.0.1:9000)
```

Key point: all Pi execution happens on VPS; local machine forwards and views only.

## Implementation steps

### 1) Install/run Pi on VPS
Run Pi as a service bound to loopback:
- host: `127.0.0.1`
- port: `7777`

Example unit: `examples/pi.service`

### 2) Run web UI/API gateway on VPS
The web layer talks to Pi runtime at `http://127.0.0.1:7777` and listens on `127.0.0.1:9000`.

Example compose: `examples/docker-compose.yml`

### 3) From local, forward a port to VPS web UI

```bash
ssh -N -L 9000:127.0.0.1:9000 user@your-vps.example
```

Open `http://localhost:9000` locally.

No Pi binary is needed on local machine.

## Optional: expose with HTTPS instead of SSH forwarding
If you want direct hostname access for multiple users, add a reverse proxy (Caddy/Nginx) on VPS with auth.

Example: `examples/Caddyfile`

## Security baseline
- Keep Pi/API/UI on `127.0.0.1` unless intentionally publishing.
- Use SSH keys; disable password login if possible.
- Restrict firewall to SSH (and maybe 443 if using public HTTPS).
- Store model/API secrets on VPS only.
- Isolate job workspaces and enforce runtime/resource limits.

## Verification
On VPS:
```bash
curl http://127.0.0.1:7777/health
curl http://127.0.0.1:9000/health
```

From local (with SSH tunnel up):
```bash
curl http://127.0.0.1:9000/health
```

## Failure modes
- Tunnel connects but UI unavailable: forwarded wrong remote port.
- UI loads but Pi actions fail: web gateway cannot reach `127.0.0.1:7777`.
- Works briefly then drops: set SSH keepalive options.

## Bottom line
Because Pi must exist only on VPS, use a **VPS-only Pi service** plus **local SSH `-L` forwarding** to the VPS web endpoint. This gives local browser usability without local Pi installation.
