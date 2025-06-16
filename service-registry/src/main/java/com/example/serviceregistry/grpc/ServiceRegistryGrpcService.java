package com.example.serviceregistry.grpc;

import com.example.serviceregistry.ServiceRegistryApplication;
import com.example.serviceregistry.model.ServiceInstance;
import com.example.serviceregistry.service.ServiceRegistryService;
import io.grpc.stub.StreamObserver;
import lombok.extern.slf4j.Slf4j;
import net.devh.boot.grpc.server.service.GrpcService;

import java.util.Collection;
import java.util.List;
import java.util.stream.Collectors;

@GrpcService
@Slf4j
public class ServiceRegistryGrpcService extends ServiceRegistryGrpc.ServiceRegistryImplBase{
    private final ServiceRegistryService serviceRegistryService;

    public ServiceRegistryGrpcService(ServiceRegistryService serviceRegistryService, ServiceRegistryApplication serviceRegistryApplication) {
        this.serviceRegistryService = serviceRegistryService;
    }

    @Override
    public void getHealthyServices(GetHealthyServicesRequest request, StreamObserver<GetHealthyServicesResponse> responseObserver) {
        log.info("Received gRPC request for healthy services");
        Collection<ServiceInstance> healthyServices = serviceRegistryService.getHealthyServices();
        List<GrpcServiceInstance> grpcInstances = healthyServices.stream()
                .map(instance -> GrpcServiceInstance.newBuilder()
                            .setId(instance.getId())
                            .setServiceName(instance.getServiceName())
                            .setHost(instance.getHost())
                            .setPort(instance.getPort())
                            .setUrl(instance.getUrl())
                            .setHealthPath(instance.getHealthPath())
                            .build())
                .collect(Collectors.toList());
        GetHealthyServicesResponse response = GetHealthyServicesResponse.newBuilder()
                .addAllServices(grpcInstances)
                .build();

        responseObserver.onNext(response);
        responseObserver.onCompleted();
    }

    @Override
    public void registerService(RegisterServiceRequest request, StreamObserver<ServiceRegistryResponse> responseObserver) {
        GrpcServiceInstance grpcInstance = request.getInstance();
        ServiceInstance serviceInstance = new ServiceInstance(
                grpcInstance.getId(),
                grpcInstance.getServiceName(),
                grpcInstance.getHost(),
                grpcInstance.getPort(),
                grpcInstance.getUrl(),
                grpcInstance.getHealthPath(),
                null,
                true
        );

        try {
            serviceRegistryService.registerService(serviceInstance);
            ServiceRegistryResponse response = ServiceRegistryResponse.newBuilder()
                    .setSuccess(true)
                    .setMessage("Service registered successfully")
                    .build();
            responseObserver.onNext(response);
            responseObserver.onCompleted();
        }
        catch (Exception e) {
            log.error("Failed to register service via gRPC: {}", e.getMessage());
            ServiceRegistryResponse response = ServiceRegistryResponse.newBuilder()
                    .setSuccess(false)
                    .setMessage("Failed to register service: " + e.getMessage())
                    .build();
            responseObserver.onNext(response);
            responseObserver.onError(e);
        }
    }

    @Override
    public void deregisterService(DeregisterServiceRequest request, StreamObserver<ServiceRegistryResponse> responseObserver) {
        try {
            serviceRegistryService.deregisterService(request.getInstanceId());
            ServiceRegistryResponse response = ServiceRegistryResponse.newBuilder()
                    .setSuccess(true)
                    .setMessage("Service deregistered successfully")
                    .build();
            responseObserver.onNext(response);
            responseObserver.onCompleted();
        }
        catch (Exception e) {
            log.error("Failed to deregister service via gRPC: {}", e.getMessage());
            ServiceRegistryResponse response = ServiceRegistryResponse.newBuilder()
                    .setSuccess(true)
                    .setMessage("Failed to deregister service: " + e.getMessage())
                    .build();
            responseObserver.onNext(response);
            responseObserver.onError(e);
        }
    }

    @Override
    public void sendHeartbeat(SendHeartbeatRequest request, StreamObserver<ServiceRegistryResponse> responseObserver) {
        try {
            serviceRegistryService.sendHeartbeat(request.getInstanceId());
            ServiceRegistryResponse response = ServiceRegistryResponse.newBuilder()
                    .setSuccess(true)
                    .setMessage("Heartbeat received.")
                    .build();
            responseObserver.onNext(response);
            responseObserver.onCompleted();
        }
        catch (Exception e) {
            log.error("Failed to process heartbeat via gRPC: {}", e.getMessage());
            ServiceRegistryResponse response = ServiceRegistryResponse.newBuilder()
                    .setSuccess(false)
                    .setMessage("Failed to process heartbeat: " + e.getMessage())
                    .build();
            responseObserver.onNext(response);
            responseObserver.onError(e);
        }
    }
}
