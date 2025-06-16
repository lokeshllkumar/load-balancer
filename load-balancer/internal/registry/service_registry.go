package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	pb "github.com/lokeshllkumar/load-balancer/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ServiceInstance struct {
	ID         string `json:"id"`
	URL        string `json:"url"`
	HealthPath string `json:"healthPath"`
}

type ServiceRegistryClient interface {
	GetServices() ([]ServiceInstance, error)
}

// creates HTTP or gRPC registry client
func NewServiceRegistryClient(clientType string, address string) (ServiceRegistryClient, error) {
	switch strings.ToLower(clientType) {
	case "http":
		return NewHTTPRegistryClient(address), nil
	case "grpc":
		return NewGRPCRegistryClient(address)
	default:
		return nil, fmt.Errorf("unsupported service registry client type: %s", clientType)
	}
}

type HTTPRegistryClient struct {
	registryURL string
	httpClient  *http.Client
}

type GRPCRegistryClient struct {
	conn *grpc.ClientConn
	client pb.ServiceRegistryClient
}

func NewHTTPRegistryClient(url string) *HTTPRegistryClient {
	return &HTTPRegistryClient{
		registryURL: url,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// fetch list of healthy services from the service registry via HTTP
func (c *HTTPRegistryClient) GetServices() ([]ServiceInstance, error) {
	fetchURL := c.registryURL
	// default scheme is http
	if !strings.HasPrefix(fetchURL, "http://") && !strings.HasPrefix(fetchURL, "https://") {
		fetchURL = "http://" + fetchURL
	}

	resp, err := c.httpClient.Get(fmt.Sprintf("%s/api/v1/services", fetchURL))
	if err != nil {
		return nil, fmt.Errorf("faield to conenct to HTTP service registry at %s: %w", fetchURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP service registry returned non-OK status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var services []ServiceInstance
	if err := json.NewDecoder(resp.Body).Decode(&services); err != nil {
		return nil, fmt.Errorf("failed to decode service list from HTTP registry: %w", err)
	}
	return services, nil
}

func NewGRPCRegistryClient(address string) (*GRPCRegistryClient, error) {
	// host:port prefix in address is to be used instead of "http://"
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials())) // unencrypted and unauthenticated gRPC connection to avoid need for TLS certificates, temporarily
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC client at %s: %s", address, err)
	}
	client := pb.NewServiceRegistryClient(conn)
	return &GRPCRegistryClient{
		conn: conn,
		client: client,
	}, nil
}

func (c *GRPCRegistryClient) GetServices() ([]ServiceInstance, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()

	resp, err := c.client.GetHealthyServices(ctx, &pb.GetHealthyServicesRequest{})
	if err != nil {
		return nil, fmt.Errorf("gRPC call to get healthy services failed: %w", err)
	}

	var instances []ServiceInstance
	for _, s := range resp.GetServices() {
		instances = append(instances, ServiceInstance{
			ID: s.Id,
			URL: s.Url,
			HealthPath: s.HealthPath,
		})
	}
	return instances, nil
}

func (c *GRPCRegistryClient) Close() error {
	if c.conn != nil {
		log.Println("Closing gRPC registry client connection")
		return c.conn.Close()
	}
	return nil
}