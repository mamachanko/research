# Notes: Can a CF-Deployed Spring Boot App Redeploy Itself?

## Question

Can a Spring Boot fat JAR deployed to Cloud Foundry via the CF API redeploy itself — essentially cloning itself but with different properties, profiles, or service bindings?

## Initial Thinking

The core question breaks into several sub-questions:

1. **Can an app talk to the CF API from within CF?** — Yes, the CF API is just a REST API. If the app has credentials and network access to the CF controller, it can make API calls.

2. **Can it avoid re-uploading its own JAR?** — This is the interesting part. Re-uploading a fat JAR (often 50-100MB+) from within CF would be slow and wasteful. The CF v3 API has concepts of "packages" and "droplets" that might allow reuse.

3. **How to set different config?** — CF environment variables are the standard way to configure Spring Boot apps. `SPRING_PROFILES_ACTIVE`, individual properties, or service bindings.

## Research Areas

- CF v3 API: packages, droplets, deployments
- cf-java-client library (cloudfoundry/cf-java-client)
- App filesystem layout on CF (where is the JAR?)
- Authentication options (user-provided credentials, CF_HOME, service accounts)
- VCAP_APPLICATION metadata for self-discovery

## Key Findings from Research

### CF V3 API — App Creation Workflow

The V3 API breaks deployment into discrete steps:
1. `POST /v3/apps` — create empty app shell
2. `POST /v3/packages` — create a package (type: bits or docker)
3. `POST /v3/packages/:guid/upload` — upload source bits
4. `POST /v3/builds` — stage the package into a droplet
5. `PATCH /v3/apps/:guid/relationships/current_droplet` — assign the droplet
6. `POST /v3/apps/:guid/actions/start` — start the app

### The Droplet Copy Trick

The `cf-java-client` Droplets interface has a `copy(CopyDropletRequest)` method that hits `POST /v3/droplets?source_guid=:guid`. This copies an existing droplet to a new app — **no re-upload, no re-staging needed**. This is the key to efficient self-redeployment.

Similarly, packages can be copied between apps.

### cf-java-client Integration

- Maven coords: `org.cloudfoundry:cloudfoundry-client-reactor` and `org.cloudfoundry:cloudfoundry-operations`
- Key class: `ReactorCloudFoundryClient` — low-level V3 API access
- Key class: `DefaultCloudFoundryOperations` — high-level operations (push, scale, etc.)
- Spring Boot auto-configuration available via Spring Cloud Cloud Foundry connector
- Fully reactive (Project Reactor based)

### VCAP_APPLICATION for Self-Discovery

The running app can read `VCAP_APPLICATION` to discover:
- `application_id` — its own app GUID (needed to find its droplet)
- `application_name` — its name
- `space_id` — the space it lives in
- `cf_api` — the CF API endpoint (no need to hardcode!)
- `application_uris` — its routes

### Authentication Options

1. **Environment variables** — inject CF_USERNAME/CF_PASSWORD or a client credentials pair
2. **User-provided service** — bind a service instance containing CF credentials
3. **CF API token** — pass an OAuth token directly (less secure, tokens expire)
4. **UAA client credentials** — create a UAA client with `cloud_controller.admin` or scoped permissions

### Setting Different Properties on the Clone

Via the CF V3 API, environment variables can be set per-app:
- `PATCH /v3/apps/:guid/environment_variables` to set `SPRING_PROFILES_ACTIVE`, database URLs, etc.
- Different services can be bound via `POST /v3/service_credential_bindings`
- Different routes via the routes API
- Different memory/disk via app scaling

### Practical Concerns Investigated

- **The original app is NOT affected** — the clone is a completely separate CF app entity
- **Droplet copy avoids re-staging** — saves minutes of buildpack execution time
- **Memory quotas** — the clone needs its own memory allocation within the org quota
- **Route conflicts** — the clone needs a unique route
- **Circular dependency risk** — if the clone deploys more clones, you get a fork bomb. Guard against this!

## Approaches Compared

| Approach | Re-upload JAR? | Re-stage? | Complexity |
|----------|---------------|-----------|------------|
| Upload JAR from filesystem | Yes (slow) | Yes | Medium |
| Copy package, then build | No | Yes (minutes) | Medium |
| **Copy droplet** | **No** | **No** | **Low** |
| Reference same droplet GUID | No | No | Lowest |

The droplet copy approach is the clear winner for self-redeployment.
