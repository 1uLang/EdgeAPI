package services

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/shared"
	"github.com/iwind/TeaGo/maps"
)

// OriginService 源站相关管理
type OriginService struct {
	BaseService
}

// CreateOrigin 创建源站
func (this *OriginService) CreateOrigin(ctx context.Context, req *pb.CreateOriginRequest) (*pb.CreateOriginResponse, error) {
	adminId, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	if req.Addr == nil {
		return nil, errors.New("'addr' can not be nil")
	}
	addrMap := maps.Map{
		"protocol":  req.Addr.Protocol,
		"portRange": req.Addr.PortRange,
		"host":      req.Addr.Host,
	}

	tx := this.NullTx()

	// 校验参数
	var connTimeout = &shared.TimeDuration{}
	if len(req.ConnTimeoutJSON) > 0 {
		err = json.Unmarshal(req.ConnTimeoutJSON, connTimeout)
		if err != nil {
			return nil, err
		}
	}

	var readTimeout = &shared.TimeDuration{}
	if len(req.ReadTimeoutJSON) > 0 {
		err = json.Unmarshal(req.ReadTimeoutJSON, readTimeout)
		if err != nil {
			return nil, err
		}
	}

	var idleTimeout = &shared.TimeDuration{}
	if len(req.IdleTimeoutJSON) > 0 {
		err = json.Unmarshal(req.IdleTimeoutJSON, idleTimeout)
		if err != nil {
			return nil, err
		}
	}

	originId, err := models.SharedOriginDAO.CreateOrigin(tx, adminId, userId, req.Name, string(addrMap.AsJSON()), req.Description, req.Weight, req.IsOn, connTimeout, readTimeout, idleTimeout, req.MaxConns, req.MaxIdleConns, req.Domains)
	if err != nil {
		return nil, err
	}

	return &pb.CreateOriginResponse{OriginId: originId}, nil
}

// UpdateOrigin 修改源站
func (this *OriginService) UpdateOrigin(ctx context.Context, req *pb.UpdateOriginRequest) (*pb.RPCSuccess, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		// TODO 校验权限
	}
	if req.Addr == nil {
		return nil, errors.New("'addr' can not be nil")
	}
	addrMap := maps.Map{
		"protocol":  req.Addr.Protocol,
		"portRange": req.Addr.PortRange,
		"host":      req.Addr.Host,
	}

	tx := this.NullTx()

	// 校验参数
	var connTimeout = &shared.TimeDuration{}
	if len(req.ConnTimeoutJSON) > 0 {
		err = json.Unmarshal(req.ConnTimeoutJSON, connTimeout)
		if err != nil {
			return nil, err
		}
	}

	var readTimeout = &shared.TimeDuration{}
	if len(req.ReadTimeoutJSON) > 0 {
		err = json.Unmarshal(req.ReadTimeoutJSON, readTimeout)
		if err != nil {
			return nil, err
		}
	}

	var idleTimeout = &shared.TimeDuration{}
	if len(req.IdleTimeoutJSON) > 0 {
		err = json.Unmarshal(req.IdleTimeoutJSON, idleTimeout)
		if err != nil {
			return nil, err
		}
	}

	err = models.SharedOriginDAO.UpdateOrigin(tx, req.OriginId, req.Name, string(addrMap.AsJSON()), req.Description, req.Weight, req.IsOn, connTimeout, readTimeout, idleTimeout, req.MaxConns, req.MaxIdleConns, req.Domains)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// FindEnabledOrigin 查找单个源站信息
func (this *OriginService) FindEnabledOrigin(ctx context.Context, req *pb.FindEnabledOriginRequest) (*pb.FindEnabledOriginResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		// TODO 校验权限
	}

	tx := this.NullTx()

	origin, err := models.SharedOriginDAO.FindEnabledOrigin(tx, req.OriginId)
	if err != nil {
		return nil, err
	}

	if origin == nil {
		return &pb.FindEnabledOriginResponse{Origin: nil}, nil
	}

	addr, err := origin.DecodeAddr()
	if err != nil {
		return nil, err
	}

	result := &pb.Origin{
		Id:   int64(origin.Id),
		IsOn: origin.IsOn == 1,
		Name: origin.Name,
		Addr: &pb.NetworkAddress{
			Protocol:  addr.Protocol.String(),
			Host:      addr.Host,
			PortRange: addr.PortRange,
		},
		Description: origin.Description,
		Domains:     origin.DecodeDomains(),
	}
	return &pb.FindEnabledOriginResponse{Origin: result}, nil
}

// FindEnabledOriginConfig 查找源站配置
func (this *OriginService) FindEnabledOriginConfig(ctx context.Context, req *pb.FindEnabledOriginConfigRequest) (*pb.FindEnabledOriginConfigResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		// TODO 校验权限
	}

	tx := this.NullTx()

	config, err := models.SharedOriginDAO.ComposeOriginConfig(tx, req.OriginId, nil)
	if err != nil {
		return nil, err
	}

	configData, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}

	return &pb.FindEnabledOriginConfigResponse{OriginJSON: configData}, nil
}
