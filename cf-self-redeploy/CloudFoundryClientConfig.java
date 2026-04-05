package com.example.selfdeploy;

import org.cloudfoundry.client.CloudFoundryClient;
import org.cloudfoundry.client.v3.applications.*;
import org.cloudfoundry.client.v3.droplets.*;
import org.cloudfoundry.client.v3.applications.SetApplicationCurrentDropletRequest;
import org.cloudfoundry.client.v3.applications.SetApplicationCurrentDropletResponse;
import org.cloudfoundry.reactor.ConnectionContext;
import org.cloudfoundry.reactor.DefaultConnectionContext;
import org.cloudfoundry.reactor.TokenProvider;
import org.cloudfoundry.reactor.client.ReactorCloudFoundryClient;
import org.cloudfoundry.reactor.tokenprovider.PasswordGrantTokenProvider;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

/**
 * Spring configuration that wires up the CF Java Client.
 *
 * Expects these environment variables (set via cf set-env or a user-provided service):
 *   CF_API_HOST   — e.g. api.cf.example.com
 *   CF_USERNAME   — a CF user or service account
 *   CF_PASSWORD   — its password
 *   CF_ORG        — target org
 *   CF_SPACE      — target space
 */
@Configuration
public class CloudFoundryClientConfig {

    @Bean
    DefaultConnectionContext connectionContext(
            @Value("${CF_API_HOST}") String apiHost) {
        return DefaultConnectionContext.builder()
                .apiHost(apiHost)
                .build();
    }

    @Bean
    PasswordGrantTokenProvider tokenProvider(
            @Value("${CF_USERNAME}") String username,
            @Value("${CF_PASSWORD}") String password) {
        return PasswordGrantTokenProvider.builder()
                .username(username)
                .password(password)
                .build();
    }

    @Bean
    ReactorCloudFoundryClient cloudFoundryClient(
            ConnectionContext connectionContext,
            TokenProvider tokenProvider) {
        return ReactorCloudFoundryClient.builder()
                .connectionContext(connectionContext)
                .tokenProvider(tokenProvider)
                .build();
    }
}
