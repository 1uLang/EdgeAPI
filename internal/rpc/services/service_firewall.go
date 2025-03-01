// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/stats"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/types"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"strconv"
	"time"
)

// FirewallService 防火墙全局服务
type FirewallService struct {
	BaseService
}

// ComposeFirewallGlobalBoard 组合看板数据
func (this *FirewallService) ComposeFirewallGlobalBoard(ctx context.Context, req *pb.ComposeFirewallGlobalBoardRequest) (*pb.ComposeFirewallGlobalBoardResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var now = time.Now()
	var day = timeutil.Format("Ymd")
	var w = types.Int(timeutil.Format("w"))
	if w == 0 {
		w = 7
	}
	weekFrom := timeutil.Format("Ymd", now.AddDate(0, 0, -w+1))
	weekTo := timeutil.Format("Ymd", now.AddDate(0, 0, -w+7))

	var result = &pb.ComposeFirewallGlobalBoardResponse{}
	var tx = this.NullTx()

	countDailyLog, err := stats.SharedServerHTTPFirewallDailyStatDAO.SumDailyCount(tx, 0, 0, "log", day, day)
	if err != nil {
		return nil, err
	}
	result.CountDailyLogs = countDailyLog

	countDailyBlock, err := stats.SharedServerHTTPFirewallDailyStatDAO.SumDailyCount(tx, 0, 0, "block", day, day)
	if err != nil {
		return nil, err
	}
	result.CountDailyBlocks = countDailyBlock

	countDailyCaptcha, err := stats.SharedServerHTTPFirewallDailyStatDAO.SumDailyCount(tx, 0, 0, "captcha", day, day)
	if err != nil {
		return nil, err
	}
	result.CountDailyCaptcha = countDailyCaptcha

	countWeeklyBlock, err := stats.SharedServerHTTPFirewallDailyStatDAO.SumDailyCount(tx, 0, 0, "block", weekFrom, weekTo)
	if err != nil {
		return nil, err
	}
	result.CountWeeklyBlocks = countWeeklyBlock

	// 24小时趋势
	var hourFrom = timeutil.Format("YmdH", time.Now().Add(-23*time.Hour))
	var hourTo = timeutil.Format("YmdH")
	hours, err := utils.RangeHours(hourFrom, hourTo)
	if err != nil {
		return nil, err
	}
	{
		statList, err := stats.SharedServerHTTPFirewallHourlyStatDAO.FindHourlyStats(tx, 0, 0, "log", hourFrom, hourTo)
		if err != nil {
			return nil, err
		}
		m := map[string]int64{} // day => count
		for _, stat := range statList {
			m[stat.Hour] = int64(stat.Count)
		}
		for _, hour := range hours {
			result.HourlyStats = append(result.HourlyStats, &pb.ComposeFirewallGlobalBoardResponse_HourlyStat{Hour: hour, CountLogs: m[hour]})
		}
	}
	{
		statList, err := stats.SharedServerHTTPFirewallHourlyStatDAO.FindHourlyStats(tx, 0, 0, "captcha", hourFrom, hourTo)
		if err != nil {
			return nil, err
		}
		m := map[string]int64{} // day => count
		for _, stat := range statList {
			m[stat.Hour] = int64(stat.Count)
		}
		for index, hour := range hours {
			result.HourlyStats[index].CountCaptcha = m[hour]
		}
	}
	{
		statList, err := stats.SharedServerHTTPFirewallHourlyStatDAO.FindHourlyStats(tx, 0, 0, "block", hourFrom, hourTo)
		if err != nil {
			return nil, err
		}
		m := map[string]int64{} // day => count
		for _, stat := range statList {
			m[stat.Hour] = int64(stat.Count)
		}
		for index, hour := range hours {
			result.HourlyStats[index].CountBlocks = m[hour]
		}
	}

	// 14天趋势
	dayFrom := timeutil.Format("Ymd", now.AddDate(0, 0, -14))
	days, err := utils.RangeDays(dayFrom, day)
	if err != nil {
		return nil, err
	}
	{
		statList, err := stats.SharedServerHTTPFirewallDailyStatDAO.FindDailyStats(tx, 0, 0, "log", dayFrom, day)
		if err != nil {
			return nil, err
		}
		m := map[string]int64{} // day => count
		for _, stat := range statList {
			m[stat.Day] = int64(stat.Count)
		}
		for _, day := range days {
			result.DailyStats = append(result.DailyStats, &pb.ComposeFirewallGlobalBoardResponse_DailyStat{Day: day, CountLogs: m[day]})
		}
	}
	{
		statList, err := stats.SharedServerHTTPFirewallDailyStatDAO.FindDailyStats(tx, 0, 0, "captcha", dayFrom, day)
		if err != nil {
			return nil, err
		}
		m := map[string]int64{} // day => count
		for _, stat := range statList {
			m[stat.Day] = int64(stat.Count)
		}
		for index, day := range days {
			result.DailyStats[index].CountCaptcha = m[day]
		}
	}
	{
		statList, err := stats.SharedServerHTTPFirewallDailyStatDAO.FindDailyStats(tx, 0, 0, "block", dayFrom, day)
		if err != nil {
			return nil, err
		}
		m := map[string]int64{} // day => count
		for _, stat := range statList {
			m[stat.Day] = int64(stat.Count)
		}
		for index, day := range days {
			result.DailyStats[index].CountBlocks = m[day]
		}
	}

	// 规则分组
	groupStats, err := stats.SharedServerHTTPFirewallDailyStatDAO.GroupDailyCount(tx, 0, 0, dayFrom, day, 0, 10)
	if err != nil {
		return nil, err
	}
	for _, stat := range groupStats {
		ruleGroupName, err := models.SharedHTTPFirewallRuleGroupDAO.FindHTTPFirewallRuleGroupName(tx, int64(stat.HttpFirewallRuleGroupId))
		if err != nil {
			return nil, err
		}
		if len(ruleGroupName) == 0 {
			continue
		}

		result.HttpFirewallRuleGroups = append(result.HttpFirewallRuleGroups, &pb.ComposeFirewallGlobalBoardResponse_HTTPFirewallRuleGroupStat{
			HttpFirewallRuleGroup: &pb.HTTPFirewallRuleGroup{Id: int64(stat.HttpFirewallRuleGroupId), Name: ruleGroupName},
			Count:                 int64(stat.Count),
		})
	}

	return result, nil
}

// NotifyHTTPFirewallEvent 发送告警(notify)消息
func (this *FirewallService) NotifyHTTPFirewallEvent(ctx context.Context, req *pb.NotifyHTTPFirewallEventRequest) (*pb.RPCSuccess, error) {
	nodeId, err := this.ValidateNode(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	clusterId, err := models.SharedNodeDAO.FindNodeClusterId(tx, nodeId)
	if err != nil {
		return nil, err
	}
	if clusterId <= 0 {
		return this.Success()
	}
	clusterName, err := models.SharedNodeClusterDAO.FindNodeClusterName(tx, clusterId)
	if err != nil {
		return nil, err
	}

	nodeName, err := models.SharedNodeDAO.FindNodeName(tx, nodeId)
	if err != nil {
		return nil, err
	}

	serverName, err := models.SharedServerDAO.FindEnabledServerName(tx, req.ServerId)
	if err != nil {
		return nil, err
	}

	ruleGroupName, err := models.SharedHTTPFirewallRuleGroupDAO.FindHTTPFirewallRuleGroupName(tx, req.HttpFirewallRuleGroupId)
	if err != nil {
		return nil, err
	}

	ruleSetName, err := models.SharedHTTPFirewallRuleSetDAO.FindHTTPFirewallRuleSetName(tx, req.HttpFirewallRuleSetId)
	if err != nil {
		return nil, err
	}

	msg := "集群：" + clusterName + "（ID：" + strconv.FormatInt(clusterId, 10) + "）" +
		"\n节点：" + nodeName + "（ID：" + strconv.FormatInt(nodeId, 10) + "）" +
		"\n服务：" + serverName + "（ID：" + strconv.FormatInt(req.ServerId, 10) + "）" +
		"\n规则分组：" + ruleGroupName +
		"\n规则集：" + ruleSetName +
		"\n时间：" + timeutil.FormatTime("Y-m-d H:i:s", req.CreatedAt)
	err = models.SharedMessageTaskDAO.CreateMessageTasks(tx, models.MessageTaskTarget{
		ClusterId: clusterId,
		NodeId:    nodeId,
		ServerId:  req.ServerId,
	}, models.MessageTypeFirewallEvent, "发生防火墙事件", msg)
	if err != nil {
		return nil, err
	}
	return this.Success()
}
