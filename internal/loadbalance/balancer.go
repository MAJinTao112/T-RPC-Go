package loadbalance

import "T-RPC-Go/internal/registry"

type LoadBalancer interface {
	Select([]registry.Instance) registry.Instance
}
