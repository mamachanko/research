# Notes

## 2026-04-05
- Created project folder `pi-web-client-vps-port-forward` per instructions.
- Interpreted request as: design a web UI that runs on a VPS, while accessing a Raspberry Pi service over a secure tunnel/port-forward, usable from local browser.
- Planned to provide architecture options, concrete implementation steps, security hardening, and troubleshooting.
- Chosen recommended architecture: reverse SSH tunnel from Pi to VPS + reverse proxy on VPS + auth/TLS.
- Added fallback architecture: WireGuard overlay between local machine, VPS, and Pi when bidirectional private networking is preferred.
- Included minimal reference configs (`docker-compose.yml`, `Caddyfile`, `systemd` unit) to make investigation actionable.
- Will add root README index entry with UTC completion time.
- Added runnable example config files under `examples/` for Caddy, systemd tunnel service, and docker-compose.
