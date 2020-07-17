package consul

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/consul/api"
)

type ConsulService struct {
	Ip   string
	Port int
	Tag  []string
	Name string
}

func RegisterService(addr string, service *ConsulService) error {
	consulConfig := api.DefaultConfig()
	consulConfig.Address = addr
	client, err := api.NewClient(consulConfig)
	if err != nil {
		log.Println("NewClient ret err", err.Error())
		return err
	}
	reg := &api.AgentServiceRegistration{
		ID:      fmt.Sprintf("%s-%s-%d", service.Name, service.Ip, service.Port),
		Name:    service.Name,
		Tags:    service.Tag,
		Port:    service.Port,
		Address: service.Ip,
		Check: &api.AgentServiceCheck{
			Interval:                       (time.Duration(10) * time.Second).String(),
			GRPC:                           fmt.Sprintf("%s:%d/%s", service.Ip, service.Port, service.Name),
			DeregisterCriticalServiceAfter: (time.Duration(1) * time.Minute).String(),
		},
	}
	if err := client.Agent().ServiceRegister(reg); err != nil {
		log.Println("ServiceRegister ret err", err.Error())
		return err
	}
	return nil
}
