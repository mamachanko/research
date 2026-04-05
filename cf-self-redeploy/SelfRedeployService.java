package com.example.selfdeploy;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import org.cloudfoundry.client.CloudFoundryClient;
import org.cloudfoundry.client.v3.Relationship;
import org.cloudfoundry.client.v3.ToOneRelationship;
import org.cloudfoundry.client.v3.applications.*;
import org.cloudfoundry.client.v3.droplets.*;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;
import reactor.core.publisher.Mono;

import java.util.Map;

/**
 * Service that can deploy a clone of the currently-running application
 * to Cloud Foundry with different environment / Spring profiles.
 *
 * The key trick: instead of re-uploading the fat JAR or re-staging,
 * we COPY the existing app's droplet to the new app. This is fast
 * (seconds, not minutes) and avoids large file transfers.
 */
@Service
public class SelfRedeployService {

    private static final Logger log = LoggerFactory.getLogger(SelfRedeployService.class);

    private final CloudFoundryClient cfClient;
    private final String selfAppId;
    private final String spaceId;

    public SelfRedeployService(
            CloudFoundryClient cfClient,
            @Value("${vcap.application.application_id}") String selfAppId,
            @Value("${vcap.application.space_id}") String spaceId) {
        this.cfClient = cfClient;
        this.selfAppId = selfAppId;
        this.spaceId = spaceId;
    }

    /**
     * Deploy a clone of this application with the given name and environment overrides.
     *
     * @param cloneName     the CF app name for the clone
     * @param envOverrides  environment variables to set on the clone
     *                      (e.g. {"SPRING_PROFILES_ACTIVE": "worker", "DB_URL": "..."})
     * @return the GUID of the newly created app
     */
    public Mono<String> deployClone(String cloneName, Map<String, String> envOverrides) {

        log.info("Deploying clone '{}' from app {} in space {}", cloneName, selfAppId, spaceId);

        // Step 1: Create a new empty app in the same space
        return createApp(cloneName)
                .flatMap(newAppGuid -> {

                    // Step 2: Set environment variables on the new app
                    Mono<Void> setEnv = setEnvironmentVariables(newAppGuid, envOverrides);

                    // Step 3: Copy our droplet to the new app
                    Mono<String> copyDroplet = copyCurrentDropletTo(newAppGuid);

                    // Wait for env to be set, then assign the copied droplet and start
                    return setEnv
                            .then(copyDroplet)
                            .flatMap(dropletGuid -> assignDroplet(newAppGuid, dropletGuid))
                            .then(startApp(newAppGuid))
                            .thenReturn(newAppGuid);
                })
                .doOnSuccess(guid -> log.info("Clone '{}' deployed successfully with guid {}", cloneName, guid))
                .doOnError(e -> log.error("Failed to deploy clone '{}'", cloneName, e));
    }

    /**
     * Step 1: Create an empty app shell in the same space.
     */
    private Mono<String> createApp(String appName) {
        return cfClient.applicationsV3()
                .create(CreateApplicationRequest.builder()
                        .name(appName)
                        .relationships(ApplicationRelationships.builder()
                                .space(ToOneRelationship.builder()
                                        .data(Relationship.builder()
                                                .id(spaceId)
                                                .build())
                                        .build())
                                .build())
                        .build())
                .map(response -> response.getId())
                .doOnSuccess(guid -> log.info("Created app '{}' with guid {}", appName, guid));
    }

    /**
     * Step 2: Set environment variables on the new app.
     * This is how we give the clone different Spring profiles, DB URLs, etc.
     */
    private Mono<Void> setEnvironmentVariables(String appGuid, Map<String, String> envVars) {
        if (envVars == null || envVars.isEmpty()) {
            return Mono.empty();
        }

        // Build the environment variable map for the CF API
        // The V3 API uses PATCH /v3/apps/:guid/environment_variables
        return cfClient.applicationsV3()
                .updateEnvironmentVariables(UpdateApplicationEnvironmentVariablesRequest.builder()
                        .applicationId(appGuid)
                        .putAllVar(envVars)
                        .build())
                .then()
                .doOnSuccess(v -> log.info("Set {} env vars on app {}", envVars.size(), appGuid));
    }

    /**
     * Step 3: Copy our current droplet to the target app.
     *
     * This is the critical optimization — instead of re-uploading the JAR
     * and re-staging (which takes minutes), we copy the already-compiled
     * droplet. The CF API handles this server-side.
     */
    private Mono<String> copyCurrentDropletTo(String targetAppGuid) {
        // First, find our own current droplet
        return cfClient.applicationsV3()
                .getCurrentDropletRelationship(GetApplicationCurrentDropletRelationshipRequest.builder()
                        .applicationId(selfAppId)
                        .build())
                .map(response -> response.getData().getId())
                .flatMap(currentDropletGuid -> {
                    log.info("Copying droplet {} to app {}", currentDropletGuid, targetAppGuid);

                    return cfClient.droplets()
                            .copy(CopyDropletRequest.builder()
                                    .sourceDropletId(currentDropletGuid)
                                    .relationships(DropletRelationships.builder()
                                            .application(ToOneRelationship.builder()
                                                    .data(Relationship.builder()
                                                            .id(targetAppGuid)
                                                            .build())
                                                    .build())
                                            .build())
                                    .build())
                            .map(CopyDropletResponse::getId);
                })
                .doOnSuccess(guid -> log.info("Droplet copied, new droplet guid: {}", guid));
    }

    /**
     * Step 4: Assign the copied droplet as the current droplet for the new app.
     */
    private Mono<Void> assignDroplet(String appGuid, String dropletGuid) {
        return cfClient.applicationsV3()
                .setCurrentDroplet(SetApplicationCurrentDropletRequest.builder()
                        .applicationId(appGuid)
                        .data(Relationship.builder()
                                .id(dropletGuid)
                                .build())
                        .build())
                .then()
                .doOnSuccess(v -> log.info("Assigned droplet {} to app {}", dropletGuid, appGuid));
    }

    /**
     * Step 5: Start the new app.
     */
    private Mono<Void> startApp(String appGuid) {
        return cfClient.applicationsV3()
                .start(StartApplicationRequest.builder()
                        .applicationId(appGuid)
                        .build())
                .then()
                .doOnSuccess(v -> log.info("Started app {}", appGuid));
    }
}
