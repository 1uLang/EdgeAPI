package stats

import (
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/rands"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"time"
)

type ServerDomainHourlyStatDAO dbs.DAO

func init() {
	dbs.OnReadyDone(func() {
		// 清理数据任务
		var ticker = time.NewTicker(time.Duration(rands.Int(24, 48)) * time.Hour)
		go func() {
			for range ticker.C {
				err := SharedServerDomainHourlyStatDAO.Clean(nil, 60) // 只保留60天
				if err != nil {
					remotelogs.Error("ServerDomainHourlyStatDAO", "clean expired data failed: "+err.Error())
				}
			}
		}()
	})
}

func NewServerDomainHourlyStatDAO() *ServerDomainHourlyStatDAO {
	return dbs.NewDAO(&ServerDomainHourlyStatDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeServerDomainHourlyStats",
			Model:  new(ServerDomainHourlyStat),
			PkName: "id",
		},
	}).(*ServerDomainHourlyStatDAO)
}

var SharedServerDomainHourlyStatDAO *ServerDomainHourlyStatDAO

func init() {
	dbs.OnReady(func() {
		SharedServerDomainHourlyStatDAO = NewServerDomainHourlyStatDAO()
	})
}

// IncreaseHourlyStat 增加统计数据
func (this *ServerDomainHourlyStatDAO) IncreaseHourlyStat(tx *dbs.Tx, clusterId int64, nodeId int64, serverId int64, domain string, hour string, bytes int64, cachedBytes int64, countRequests int64, countCachedRequests int64, countAttackRequests int64, attackBytes int64) error {
	if len(hour) != 10 {
		return errors.New("invalid hour '" + hour + "'")
	}
	err := this.Query(tx).
		Param("bytes", bytes).
		Param("cachedBytes", cachedBytes).
		Param("countRequests", countRequests).
		Param("countCachedRequests", countCachedRequests).
		Param("countAttackRequests", countAttackRequests).
		Param("attackBytes", attackBytes).
		InsertOrUpdateQuickly(maps.Map{
			"clusterId":           clusterId,
			"nodeId":              nodeId,
			"serverId":            serverId,
			"hour":                hour,
			"domain":              domain,
			"bytes":               bytes,
			"cachedBytes":         cachedBytes,
			"countRequests":       countRequests,
			"countCachedRequests": countCachedRequests,
			"countAttackRequests": countAttackRequests,
			"attackBytes":         attackBytes,
		}, maps.Map{
			"bytes":               dbs.SQL("bytes+:bytes"),
			"cachedBytes":         dbs.SQL("cachedBytes+:cachedBytes"),
			"countRequests":       dbs.SQL("countRequests+:countRequests"),
			"countCachedRequests": dbs.SQL("countCachedRequests+:countCachedRequests"),
			"countAttackRequests": dbs.SQL("countAttackRequests+:countAttackRequests"),
			"attackBytes":         dbs.SQL("attackBytes+:attackBytes"),
		})
	if err != nil {
		return err
	}
	return nil
}

// FindTopDomainStats 取得一定时间内的域名排行数据
func (this *ServerDomainHourlyStatDAO) FindTopDomainStats(tx *dbs.Tx, hourFrom string, hourTo string, size int64) (result []*ServerDomainHourlyStat, err error) {
	// TODO 节点如果已经被删除，则忽略
	_, err = this.Query(tx).
		Between("hour", hourFrom, hourTo).
		Result("domain, MIN(serverId) AS serverId, SUM(bytes) AS bytes, SUM(cachedBytes) AS cachedBytes, SUM(countRequests) AS countRequests, SUM(countCachedRequests) AS countCachedRequests, SUM(countAttackRequests) AS countAttackRequests, SUM(attackBytes) AS attackBytes").
		Group("domain").
		Desc("countRequests").
		Limit(size).
		Slice(&result).
		FindAll()
	return
}

// FindTopDomainStatsWithClusterId 取得集群上的一定时间内的域名排行数据
func (this *ServerDomainHourlyStatDAO) FindTopDomainStatsWithClusterId(tx *dbs.Tx, clusterId int64, hourFrom string, hourTo string, size int64) (result []*ServerDomainHourlyStat, err error) {
	// TODO 节点如果已经被删除，则忽略
	_, err = this.Query(tx).
		Attr("clusterId", clusterId).
		Between("hour", hourFrom, hourTo).
		Result("domain, MIN(serverId) AS serverId, SUM(bytes) AS bytes, SUM(cachedBytes) AS cachedBytes, SUM(countRequests) AS countRequests, SUM(countCachedRequests) AS countCachedRequests, SUM(countAttackRequests) AS countAttackRequests, SUM(attackBytes) AS attackBytes").
		Group("domain").
		Desc("countRequests").
		Limit(size).
		Slice(&result).
		FindAll()
	return
}

// FindTopDomainStatsWithNodeId 取得节点上的一定时间内的域名排行数据
func (this *ServerDomainHourlyStatDAO) FindTopDomainStatsWithNodeId(tx *dbs.Tx, nodeId int64, hourFrom string, hourTo string, size int64) (result []*ServerDomainHourlyStat, err error) {
	// TODO 节点如果已经被删除，则忽略
	_, err = this.Query(tx).
		Attr("nodeId", nodeId).
		Between("hour", hourFrom, hourTo).
		Result("domain, MIN(serverId) AS serverId, SUM(bytes) AS bytes, SUM(cachedBytes) AS cachedBytes, SUM(countRequests) AS countRequests, SUM(countCachedRequests) AS countCachedRequests, SUM(countAttackRequests) AS countAttackRequests, SUM(attackBytes) AS attackBytes").
		Group("domain").
		Desc("countRequests").
		Limit(size).
		Slice(&result).
		FindAll()
	return
}

// FindTopDomainStatsWithServerId 取得某个服务的一定时间内的域名排行数据
func (this *ServerDomainHourlyStatDAO) FindTopDomainStatsWithServerId(tx *dbs.Tx, serverId int64, hourFrom string, hourTo string, size int64) (result []*ServerDomainHourlyStat, err error) {
	// TODO 节点如果已经被删除，则忽略
	_, err = this.Query(tx).
		Attr("serverId", serverId).
		Between("hour", hourFrom, hourTo).
		Result("domain, MIN(serverId) AS serverId, SUM(bytes) AS bytes, SUM(cachedBytes) AS cachedBytes, SUM(countRequests) AS countRequests, SUM(countCachedRequests) AS countCachedRequests, SUM(countAttackRequests) AS countAttackRequests, SUM(attackBytes) AS attackBytes").
		Group("domain").
		Desc("countRequests").
		Limit(size).
		Slice(&result).
		FindAll()
	return
}

// Clean 清理历史数据
func (this *ServerDomainHourlyStatDAO) Clean(tx *dbs.Tx, days int) error {
	var hour = timeutil.Format("Ymd00", time.Now().AddDate(0, 0, -days))
	_, err := this.Query(tx).
		Lt("hour", hour).
		Delete()
	return err
}
