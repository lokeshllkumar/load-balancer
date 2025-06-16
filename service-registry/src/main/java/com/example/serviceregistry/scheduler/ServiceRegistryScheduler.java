package com.example.serviceregistry.scheduler;

import com.example.serviceregistry.service.ServiceRegistryService;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Component;

import java.time.LocalDateTime;

@Component
@Slf4j
public class ServiceRegistryScheduler {
    private final ServiceRegistryService serviceRegistryService;

    @Value("${service.cleanup.interval.milliseconds:5000}")
    private long cleanupIntervalMillis;
    
    public ServiceRegistryScheduler(ServiceRegistryService serviceRegistryService) {
        this.serviceRegistryService = serviceRegistryService;

    }

    // periodic cleanup
    @Scheduled(fixedRateString = "${service.cleanup.interval.milliseconds}")
    public void cleanupUnhealthyServices() {
        log.debug("Running scheduled cleanup of unhealthy services...");
        int initialHealthyCount = serviceRegistryService.getHealthyServices().size();
        

        serviceRegistryService.getHealthyServices();

        int finalHealthyCount = serviceRegistryService.getHealthyServices().size();
        if (finalHealthyCount < initialHealthyCount) {
            log.info("Cleanup completed. {} services are now healthy ({} services marked unhealthy)",
            finalHealthyCount, initialHealthyCount - finalHealthyCount);
        }
        else {
            log.debug("Cleanup completed. All {} services are healthy", finalHealthyCount);
        }
    }
}
