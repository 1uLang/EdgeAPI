// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.
// +build plus

package nodes

import (
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/services"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"google.golang.org/grpc"
)

func APIAuthorityKeyServicesRegister(node *APINode, server *grpc.Server) {
	{
		instance := node.serviceInstance(&services.AuthorityKeyService{}).(*services.AuthorityKeyService)
		pb.RegisterAuthorityKeyServiceServer(server, instance)
		node.rest(instance)
	}
}

