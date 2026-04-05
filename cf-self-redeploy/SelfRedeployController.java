package com.example.selfdeploy;

import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import reactor.core.publisher.Mono;

import java.util.Map;

/**
 * REST endpoint to trigger self-redeployment.
 *
 * POST /api/clone
 * {
 *   "name": "my-app-worker",
 *   "env": {
 *     "SPRING_PROFILES_ACTIVE": "worker",
 *     "WORKER_QUEUE": "high-priority"
 *   }
 * }
 */
@RestController
@RequestMapping("/api")
public class SelfRedeployController {

    private final SelfRedeployService redeployService;

    public SelfRedeployController(SelfRedeployService redeployService) {
        this.redeployService = redeployService;
    }

    @PostMapping("/clone")
    public Mono<ResponseEntity<Map<String, String>>> cloneApp(
            @RequestBody CloneRequest request) {

        return redeployService.deployClone(request.getName(), request.getEnv())
                .map(appGuid -> ResponseEntity.ok(Map.of(
                        "status", "deployed",
                        "appGuid", appGuid,
                        "name", request.getName()
                )));
    }

    public static class CloneRequest {
        private String name;
        private Map<String, String> env;

        public String getName() { return name; }
        public void setName(String name) { this.name = name; }
        public Map<String, String> getEnv() { return env; }
        public void setEnv(Map<String, String> env) { this.env = env; }
    }
}
