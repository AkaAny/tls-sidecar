package trust_center

import (
	"crypto/x509"
	"net/http"
)

const IdentityTypeHeaderKey = "X-Identity-Type"
const IdentityTypeCertRPC = "cert-rpc"
const IdentityTypeCertStudent = "cert-student"

type RPCIdentityHandler struct {
}

func (x *RPCIdentityHandler) HandleIdentity(r *http.Request, peerCert *x509.Certificate) (http.Header, error) {
	var respHeader = make(http.Header)
	var identityType = r.Header.Get("X-Identity-Type")
	if identityType == "" {
		respHeader.Set("X-Identity-Type", "user-anonymous")
		return respHeader, nil
	}
	switch identityType {
	case IdentityTypeCertRPC:
		serviceCertInfo, err := NewServiceCertificate(peerCert)
		if err != nil {
			return nil, err
		}
		respHeader.Set("X-From-Deploy", serviceCertInfo.DeployID)
		respHeader.Set("X-From-Service", serviceCertInfo.ServiceID)
		return respHeader, nil
	}
	return respHeader, nil
}
