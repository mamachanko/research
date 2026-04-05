# Can a Cloud Foundry Spring Boot App Redeploy Itself?

**TL;DR: Yes.** A Spring Boot app running on Cloud Foundry can use the CF V3 API to deploy a clone of itself with different properties, profiles, and service bindings — and it can do so efficiently by copying its own droplet rather than re-uploading or re-staging.

## The Question

Can a Spring Boot fat JAR that was deployed to Cloud Foundry via `cf push` (or the CF API) programmatically deploy another instance of itself to CF, but configured differently? For example, the same codebase running as a web frontend in one instance and a background worker in another, differentiated only by Spring profiles.

## Answer: Yes — Three Approaches, One Clear Winner

### Approach 1: Re-upload the JAR from the Filesystem (Slow)

The running app can access its own JAR at `/home/vcap/app/` on the CF container filesystem. It could read this file and upload it via `POST /v3/packages/:guid/upload`. This works but is **slow** — you're uploading a 50-100MB+ JAR over the network, then waiting for staging (buildpack compilation) which takes 1-3 minutes.

### Approach 2: Copy the Package, Then Build (Medium)

The CF V3 API allows copying a package from one app to another. This avoids the upload but still requires staging (building a droplet from the package), which takes 1-3 minutes.

### Approach 3: Copy the Droplet (Fast) ✓

**This is the recommended approach.** A droplet is an already-staged, ready-to-run artifact. The CF V3 API provides a droplet copy endpoint:

```
POST /v3/droplets?source_guid=<source-droplet-guid>
Body: { "relationships": { "app": { "data": { "guid": "<target-app-guid>" } } } }
```

This copies the compiled droplet server-side — **no re-upload, no re-staging**. The clone can start in seconds rather than minutes.

## How It Works — Step by Step

```
┌─────────────────────────────────────────────────┐
│  Running App ("my-app")                          │
│  VCAP_APPLICATION → { application_id, space_id } │
│  CF credentials in env vars                      │
│                                                   │
│  1. Read own app GUID from VCAP_APPLICATION       │
│  2. GET current droplet GUID via CF V3 API        │
│  3. POST /v3/apps → create new app shell          │
│  4. PATCH env vars on new app (profiles, config)  │
│  5. POST /v3/droplets?source_guid=... → copy      │
│  6. PATCH /v3/apps/:guid/current_droplet → assign │
│  7. POST /v3/apps/:guid/actions/start → launch    │
└─────────────────────────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────────────────┐
│  Clone App ("my-app-worker")                     │
│  Same code, different Spring profile             │
│  SPRING_PROFILES_ACTIVE=worker                   │
│  Different service bindings, routes, etc.        │
└─────────────────────────────────────────────────┘
```

## Implementation

The proof-of-concept uses three classes:

### 1. `CloudFoundryClientConfig.java` — CF Client Wiring

Sets up `ReactorCloudFoundryClient` using credentials from environment variables. The app needs:
- `CF_API_HOST` — the CF API endpoint (also available from `VCAP_APPLICATION.cf_api`)
- `CF_USERNAME` / `CF_PASSWORD` — a service account with SpaceDeveloper role
- `CF_ORG` / `CF_SPACE` — target org and space

### 2. `SelfRedeployService.java` — Core Logic

The service performs the five-step clone process:
1. **Create app** — `POST /v3/apps` with the target space
2. **Set env vars** — `PATCH /v3/apps/:guid/environment_variables` to configure different Spring profiles, database URLs, etc.
3. **Copy droplet** — `POST /v3/droplets?source_guid=...` to copy the compiled droplet (no re-staging!)
4. **Assign droplet** — `PATCH /v3/apps/:guid/relationships/current_droplet`
5. **Start app** — `POST /v3/apps/:guid/actions/start`

### 3. `SelfRedeployController.java` — REST Trigger

Exposes `POST /api/clone` to trigger redeployment:

```json
POST /api/clone
{
  "name": "my-app-worker",
  "env": {
    "SPRING_PROFILES_ACTIVE": "worker",
    "WORKER_QUEUE": "high-priority",
    "DB_URL": "jdbc:postgresql://worker-db:5432/tasks"
  }
}
```

## Dependencies

```xml
<dependency>
    <groupId>org.cloudfoundry</groupId>
    <artifactId>cloudfoundry-client-reactor</artifactId>
    <version>5.12.0.RELEASE</version>
</dependency>
<dependency>
    <groupId>org.cloudfoundry</groupId>
    <artifactId>cloudfoundry-operations</artifactId>
    <version>5.12.0.RELEASE</version>
</dependency>
```

The `cf-java-client` is a mature, Reactor-based library maintained by the Cloud Foundry community. It maps 1:1 to the CF V3 REST API.

## Self-Discovery via VCAP_APPLICATION

A running CF app can discover its own identity without hardcoding anything. The `VCAP_APPLICATION` environment variable (automatically injected by CF) contains:

```json
{
  "application_id": "abc-123-...",
  "application_name": "my-app",
  "space_id": "def-456-...",
  "space_name": "production",
  "cf_api": "https://api.cf.example.com",
  "application_uris": ["my-app.cf.example.com"]
}
```

Spring Boot auto-binds these via `${vcap.application.application_id}`, etc.

## Authentication

The app needs CF API credentials with the **SpaceDeveloper** role (minimum). Options:

| Method | Pros | Cons |
|--------|------|------|
| Env vars (`CF_USERNAME`/`CF_PASSWORD`) | Simple | Credentials in plain text in env |
| User-provided service | Credentials managed as a service binding | Extra setup |
| UAA client credentials grant | No user password needed, scoped permissions | Requires UAA admin to create client |

For production, UAA client credentials with scoped permissions is recommended.

## Configuring the Clone Differently

The whole point is deploying the same code with different behavior. Options:

- **Spring profiles** — Set `SPRING_PROFILES_ACTIVE=worker` to activate `application-worker.yml`
- **Individual properties** — Set `SERVER_PORT`, `DB_URL`, etc. as env vars
- **Service bindings** — Bind different database/cache/queue instances via `POST /v3/service_credential_bindings`
- **Routes** — Map to different URLs via the routes API
- **Scaling** — Set different memory/disk/instances via `PATCH /v3/apps/:guid/features` or process scaling
- **Health check** — Configure different health check types for workers vs web apps

## Practical Concerns

### Security: Prevent Fork Bombs
If the clone also has CF credentials, it could clone itself recursively. Mitigations:
- **Don't give the clone CF credentials** — omit `CF_USERNAME`/`CF_PASSWORD` from its env
- **Use a naming convention guard** — refuse to clone if app name matches a clone pattern
- **Set a `CLONE_DEPTH` env var** — increment it on each clone, refuse if > 0

### The Original App Is Safe
The clone is a completely separate CF app entity. Creating it does not affect the original — no restarts, no downtime, no shared state.

### Resource Quotas
Each clone consumes memory, disk, and app count quota in the CF org. Ensure sufficient quota before spawning clones.

### Route Management
The clone needs its own route. If you don't assign one, it won't be reachable via HTTP (which may be fine for background workers that consume from a queue).

## Use Cases

1. **Worker scaling** — Web app spawns worker instances with `SPRING_PROFILES_ACTIVE=worker` during high load
2. **Multi-tenant deployment** — Single codebase deploys per-tenant instances with different database bindings
3. **Blue-green self-deployment** — App deploys a new version of itself, switches routes, then cleans up the old instance
4. **Scheduled job runners** — App spawns short-lived task instances with specific configurations
5. **Feature branch testing** — CI deploys the base app, which then clones itself with different feature flags

## Sources

- [CF V3 API Reference](https://v3-apidocs.cloudfoundry.org/)
- [cf-java-client GitHub](https://github.com/cloudfoundry/cf-java-client)
- [How to Create an App Using V3 of the CC API](https://github.com/cloudfoundry/cloud_controller_ng/wiki/How-to-Create-an-App-Using-V3-of-the-CC-API)
- [Spring Tips: CF Java Client Autoconfiguration](https://spring.io/blog/2020/04/01/spring-tips-manipulating-the-platform-with-the-spring-cloud-cloud-foundry-java-client-autoconfiguration/)
- [CF Environment Variables Docs](https://docs.cloudfoundry.org/devguide/deploy-apps/environment-variable.html)
- [CF Java Client Library Docs](https://docs.cloudfoundry.org/buildpacks/java/java-client.html)
