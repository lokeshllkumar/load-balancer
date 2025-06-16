package com.example.serviceregistry.controller;

import com.example.serviceregistry.model.ServiceInstance;
import com.example.serviceregistry.service.ServiceRegistryService;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.security.Provider.Service;
import java.util.Collection;
import java.util.List;
import java.util.stream.Collectors;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestParam;


@RestController
@RequestMapping("/api/v1/services")
@Slf4j
public class ServiceRegistryRestController {
    private final ServiceRegistryService serviceRegistryService;

    public ServiceRegistryRestController(ServiceRegistryService serviceRegistryService) {
        this.serviceRegistryService = serviceRegistryService;
    }

    // registration endpoint
    @PostMapping("/register")
    public ResponseEntity<ServiceInstance> registerService(@RequestBody ServiceInstance instance) {
        if (instance.getId() == null || instance.getId().isEmpty() ||
        instance.getServiceName() == null || instance.getServiceName().isEmpty() ||
        instance.getUrl() == null || instance.getUrl().isEmpty() ||
        instance.getHealthPath() == null || instance.getHealthPath().isEmpty()) {
            return new ResponseEntity<>(HttpStatus.BAD_REQUEST);
        }
        ServiceInstance registered = serviceRegistryService.registerService(instance);
        return new ResponseEntity<>(registered, HttpStatus.CREATED);
    }

    // heartbeat endpoint
    @PostMapping("/heartbeat/{instanceId}")
    public ResponseEntity<Void> sendHeartbeat(@PathVariable String instanceId) {
        serviceRegistryService.sendHeartbeat(instanceId);

        // 200 OK
        return ResponseEntity.ok().build();
    }

    @DeleteMapping("/deregister/{instanceId}")
    public ResponseEntity<Void> deregisterService(@PathVariable String instanceId) {
        serviceRegistryService.deregisterService(instanceId);
        return ResponseEntity.noContent().build();
    }

    @GetMapping
    public ResponseEntity<List<ServiceInstance>> getHealthyServices() {
        List<ServiceInstance> lightWeightInstances = serviceRegistryService.getHealthyServices().stream()
                .map(instance -> new ServiceInstance(instance.getId(), instance.getServiceName(), instance.getHost(), instance.getPort(), instance.getUrl(), instance.getHealthPath(), null, false))
                .collect(Collectors.toList());
        return ResponseEntity.ok(lightWeightInstances);
    }   
}