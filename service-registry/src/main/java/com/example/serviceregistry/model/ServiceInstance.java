package com.example.serviceregistry.model;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.LocalDateTime;

@Data
@AllArgsConstructor
@NoArgsConstructor
public class ServiceInstance {
    private String id;
    private String serviceName;
    private String host;
    private int port;
    private String url;
    private String healthPath;
    private LocalDateTime lastHeartbeat;
    private boolean alive;
}
