package nrpc

import (

	//used for generated stub files

	_ "github.com/golang/protobuf/proto"
	nats "github.com/nats-io/nats.go"
)

// ServiceInfo holds name of service and name of proto
type ServiceInfo struct {
	ServiceName string
	ProtoName   string
}

// ServerService methods to invoke registartion of service
type ServerService interface {
	ServiceInfo() *ServiceInfo
	RunRegisterServerService(nc *nats.Conn, t *Trigger, h *Handler)
}

// ServiceRegistery holds all the server services written in proto file
var ServiceRegistery = NewServiceRegistry()

// ServiceRegistry data structure to hold the services
type ServiceRegistry struct {
	ServerServices map[string]ServerService
}

// RegisterServerService resgisters server services
func (sr *ServiceRegistry) RegisterServerService(service ServerService) {
	sr.ServerServices[service.ServiceInfo().ProtoName+service.ServiceInfo().ServiceName] = service
}

// NewServiceRegistry creates a new service registry
func NewServiceRegistry() *ServiceRegistry {
	return &ServiceRegistry{ServerServices: make(map[string]ServerService)}
}