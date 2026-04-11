# Notes

## 2026-04-05
- Restarted investigation due to clarified scope: user meant "Pi" as a coding agent, not Raspberry Pi hardware.
- Reframed architecture around a coding-agent runtime on VPS with browser access from local via SSH local port-forwarding.
- Chose design: agent runner API bound to loopback + web client bound to loopback + optional VPS reverse proxy for team access.
- Will include minimal reference artifacts (compose, Caddy, systemd, helper scripts) focused on secure localhost-first access.
- Added examples for loopback-only deployment with docker-compose and an optional Caddy front door.
- Added systemd service template for agent runtime and a local SSH-forward helper script.
- Focused guidance on SSH `-L` because user specifically asked for local use while running remotely on VPS.
