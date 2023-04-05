package sidecar

import (
	"context"
	"github.com/samber/lo"
	tls_sidecar "tls-sidecar"
	sidecar_proto "tls-sidecar/gen/go/proto/sidecar"
)

type BackendToSidecarServiceImpl struct {
	sidecar_proto.UnimplementedBackendToSidecarServiceServer
	routeHandler *tls_sidecar.ServiceRouteHandler
}

func NewBackendToSidecarServiceImpl(routeHandler *tls_sidecar.ServiceRouteHandler) *BackendToSidecarServiceImpl {
	return &BackendToSidecarServiceImpl{
		routeHandler: routeHandler,
	}
}

func (b *BackendToSidecarServiceImpl) ReportHTTPAPI(ctx context.Context,
	request *sidecar_proto.ReportHTTPAPIRequest) (*sidecar_proto.ReportHTTPAPIResponse, error) {
	//peerInfo, ok := peer.FromContext(ctx)
	//if !ok {
	//	return nil, status.Error(codes.FailedPrecondition, "err get peer info from context")
	//}
	//peerInfo.Addr.String()
	var httpAPIInfoLocals = lo.Map(request.ApiInfos,
		func(item *sidecar_proto.HttpAPIInfo, index int) *tls_sidecar.HttpAPIInfoLocal {
			return &tls_sidecar.HttpAPIInfoLocal{
				FullPath:              item.FullPath,
				Methods:               item.Methods,
				SupportedIdentityType: item.SupportedIdentityType,
			}
		})
	b.routeHandler.AddHttpAPIInfo(httpAPIInfoLocals)
	return &sidecar_proto.ReportHTTPAPIResponse{}, nil
}
