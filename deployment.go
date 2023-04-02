package tls_sidecar

type DeployInfo struct {
	ID string
	//http path ->
	Host             string
	IDServiceInfoMap map[string]*ServiceInfo
}

func (x *DeployInfo) GetServiceInfo(serviceID string) *ServiceInfo {
	serviceInfo, ok := x.IDServiceInfoMap[serviceID]
	if !ok {
		return nil
	}
	return serviceInfo
}
