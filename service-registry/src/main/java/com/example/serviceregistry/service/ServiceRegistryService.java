package com.example.serviceregistry.service;

import com.example.serviceregistry.model.ServiceInstance;
import io.micrometer.core.instrument.Gauge;
import io.micrometer.core.instrument.MeterRegistry;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;

import javax.annotation.PostConstruct;
import java.time.LocalDateTime;
import java.util.Collection;
import java.util.Map;
import java.util.Optional;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.atomic.AtomicInteger;
import java.util.stream.Collectors;

@Service
@Slf4j // logger
public class ServiceRegistryService {
    private final Map<String, ServiceInstance> registeredServices = new ConcurrentHashMap<>();

    @Value("${service.heartbeat.timeout.seconds:10}")
    private long heartbeatTimeoutSeconds;

    private final MeterRegistry meterRegistry;
    private final AtomicInteger registeredServicesCount;

    public ServiceRegistryService(MeterRegistry meterRegistry) {
        this.meterRegistry = meterRegistry;
        this.registeredServicesCount = meterRegistry.gauge("service_registry_registered_services_total", new AtomicInteger(0));
    }

    @PostConstruct
    public void init() {
        log.info("ServiceRegistryService initialized. Heartbeat timeout: {} seconds", heartbeatTimeoutSeconds);
    }

    // to register a service
    public ServiceInstance registerService(ServiceInstance instance) {
        instance.setLastHeartbeat(LocalDateTime.now());
        instance.setAlive(true);
        registeredServices.put(instance.getId(), instance);
        registeredServicesCount.set(registeredServices.size());
        log.info("Service registered/updated: {} (ID: {})", instance.getServiceName(), instance.getId());

        // registering a gauge for each service's heartbeat status
        Gauge.builder("service_registry_instance_alive", instance, si -> si.isAlive() ? 1.0 : 0.0)
                .tag("instance_id", instance.getId())
                .tag("service_name", instance.getServiceName())
                .description("Service instance status (0: down, 1: alive)")
                .strongReference(true)
                .register(meterRegistry);

        return instance;
    }

    // to deregister a service
    public void deregisterService(String instanceId) {
        ServiceInstance removed = registeredServices.remove(instanceId);
        if (removed != null) {
            log.info("Service deregistered: {} (ID: {})", removed.getServiceName(), removed.getId());
            registeredServicesCount.set(registeredServices.size());
            
            // GC automatically handles cleanup and removes gauge of the removed service

            // for manual cleanup
            // meterRegistry.get("service_registry_instance_alive").tag("instance_id", instanceId).tag("service_name", removed.getServiceName()).gauge().set(0.0);
        }
        else {
            log.warn("Attempted to deregister non-existent service instance with ID: {}", instanceId);
        }
    }

    // updating heartbeat for a service instance
    public void sendHeartbeat(String instanceId) {
        ServiceInstance instance = registeredServices.get(instanceId);
        if (instance != null) {
            instance.setLastHeartbeat(LocalDateTime.now());

            if (!instance.isAlive()) {
                // automatically updates gauge value to 1
                instance.setAlive(true);
                log.info("Service {} (ID: {}) is now healthy via heartbeat", instance.getServiceName(), instance.getId());

                // manual updation
                // meterRegistry.get("service_registry_instance_alive").tag("instance_id", instanceId).tag("service_name", instance.getServiceName()).gauge().set(1.0);
            }
        }
        else {
            log.warn("Heartbeat received for unkown service instance");
        }
    }

    public Collection<ServiceInstance> getHealthyServices() {
        return registeredServices.values().stream()
                .filter(instance -> {
                    // check if the heartbeat is within the timeout
                    boolean isHealthy = instance.getLastHeartbeat().plusSeconds(heartbeatTimeoutSeconds).isAfter(LocalDateTime.now());
                    // checking if the previously alive heartbeat expired
                    if (!isHealthy & instance.isAlive()) {
                        instance.setAlive(false);
                        log.warn("Service: {} (ID: {}) marked unhealthy due to an expired heartbeat", instance.getServiceName(), instance.getId());
                        // manual updation
                        //meterRegistry.get("service_registry_instance_alive").tag("instance_id", instance.getId()).tag("service_name", instance.getServiceName()).gauge().set(1.0);
                    }
                    else if (isHealthy && !instance.isAlive()) {
                        instance.setAlive(true);
                        log.info("Service {} (ID: {}) is now healthy (due to fresh heartbeat).", instance.getServiceName(), instance.getId());
                        // manual updation
                        // meterRegistry.get("service_registry_instance_alive").tag("instance_id", instance.getId()).tag("service_name", instance.getServiceName()).gauge().set(1.0);
                    }
                    return isHealthy;
                })
                .collect(Collectors.toList());
    }

    // retrieves a service instance by its ID
    public Optional<ServiceInstance> getServiceInstance(String instanceId) {
        return Optional.ofNullable(registeredServices.get(instanceId));
    }
}