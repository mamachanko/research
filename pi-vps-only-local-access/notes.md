# Notes

## 2026-04-05
- Clarification: Pi cannot be installed on local machine; Pi must run only on VPS.
- Adjusted architecture accordingly: local device only needs browser + SSH client.
- Plan: document VPS-only runtime, local SSH -L access, and optional public HTTPS mode.
- Will include practical examples: systemd unit for Pi on VPS, Caddy reverse proxy, compose, and SSH helper.
- Wrote report emphasizing VPS-only Pi runtime and local client-only usage.
- Added service and deployment examples with loopback binding to avoid accidental public exposure.
- Included explicit statement that local machine does not run Pi at all.
