package loadbalance

import "github.com/MAJinTao112/T-RPC-Go/internal/registry"

type LoadBalancer interface {
	Select([]registry.Instance) registry.Instance
}
