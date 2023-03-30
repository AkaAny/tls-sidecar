package trust_center

import "crypto/x509"

type TrustedServiceCertificate struct {
	ServiceID   string
	DeployID    string
	Certificate *x509.Certificate
}

func NewServiceCertificate(cert *x509.Certificate) (*TrustedServiceCertificate, error) {
	//u trust deploy's CA and it contains all u need
	var serviceID = cert.Subject.CommonName
	var deployID = cert.Subject.OrganizationalUnit[0]
	return &TrustedServiceCertificate{
		ServiceID:   serviceID,
		DeployID:    deployID,
		Certificate: cert,
	}, nil
}
