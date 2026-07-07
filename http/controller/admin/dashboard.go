package admin

import (
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lejianwen/rustdesk-api/v2/http/response"
	"github.com/lejianwen/rustdesk-api/v2/model"
	"github.com/lejianwen/rustdesk-api/v2/service"
)

type Dashboard struct {
}

type DashboardSummary struct {
	Stats              DashboardStats          `json:"stats"`
	Trend              []DashboardTrendPoint   `json:"trend"`
	DeviceDistribution []DashboardDistribution `json:"device_distribution"`
	LoginLogs          []DashboardLoginLog     `json:"login_logs"`
	Devices            []DashboardDevice       `json:"devices"`
	Services           []DashboardService      `json:"services"`
	Uptime             string                  `json:"uptime"`
}

type DashboardStats struct {
	TotalDevices      int64 `json:"total_devices"`
	OnlineDevices     int64 `json:"online_devices"`
	OfflineDevices    int64 `json:"offline_devices"`
	WarningDevices    int64 `json:"warning_devices"`
	TodayConnections  int64 `json:"today_connections"`
	ActiveUsers       int64 `json:"active_users"`
	WarningEvents     int64 `json:"warning_events"`
	DeviceGroups      int64 `json:"device_groups"`
	WeekShareRecords  int64 `json:"week_share_records"`
	UnassignedDevices int64 `json:"unassigned_devices"`
}

type DashboardTrendPoint struct {
	Date        string `json:"date"`
	Connections int    `json:"connections"`
	Devices     int    `json:"devices"`
}

type DashboardDistribution struct {
	Label   string `json:"label"`
	Value   int64  `json:"value"`
	Percent string `json:"percent"`
	Color   string `json:"color"`
}

type DashboardLoginLog struct {
	Time   string `json:"time"`
	User   string `json:"user"`
	IP     string `json:"ip"`
	Device string `json:"device"`
	Result string `json:"result"`
}

type DashboardDevice struct {
	OS             string `json:"os"`
	Name           string `json:"name"`
	ID             string `json:"id"`
	Status         string `json:"status"`
	Group          string `json:"group"`
	LastOnlineTime int64  `json:"last_online_time"`
}

type DashboardService struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

// Summary 仪表盘统计
// @Tags 仪表盘
// @Summary 仪表盘统计
// @Description 仪表盘统计
// @Accept  json
// @Produce  json
// @Success 200 {object} response.Response{data=DashboardSummary}
// @Failure 500 {object} response.Response
// @Router /admin/dashboard/summary [get]
// @Security token
func (d *Dashboard) Summary(c *gin.Context) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	weekStart := today.AddDate(0, 0, -6)
	onlineAfter := now.Unix() - 60

	stats := DashboardStats{}
	db := service.DB
	db.Model(&model.Peer{}).Count(&stats.TotalDevices)
	db.Model(&model.Peer{}).Where("last_online_time >= ?", onlineAfter).Count(&stats.OnlineDevices)
	stats.OfflineDevices = stats.TotalDevices - stats.OnlineDevices
	db.Model(&model.Peer{}).Where("group_id = ?", 0).Count(&stats.UnassignedDevices)
	db.Model(&model.AuditConn{}).Where("created_at >= ?", today).Count(&stats.TodayConnections)
	db.Model(&model.AuditConn{}).Where("close_time = ? and created_at < ?", 0, now.Add(-10*time.Minute)).Count(&stats.WarningEvents)
	db.Model(&model.DeviceGroup{}).Count(&stats.DeviceGroups)
	db.Model(&model.ShareRecord{}).Where("created_at >= ?", weekStart).Count(&stats.WeekShareRecords)
	db.Model(&model.LoginLog{}).
		Where("created_at >= ? and is_deleted = ?", weekStart, model.IsDeletedNo).
		Distinct("user_id").
		Count(&stats.ActiveUsers)

	summary := DashboardSummary{
		Stats:              stats,
		Trend:              d.trend(weekStart, today),
		DeviceDistribution: d.deviceDistribution(stats),
		LoginLogs:          d.loginLogs(),
		Devices:            d.devices(onlineAfter),
		Services:           d.services(),
		Uptime:             d.uptime(),
	}
	response.Success(c, summary)
}

func (d *Dashboard) trend(weekStart, today time.Time) []DashboardTrendPoint {
	points := make([]DashboardTrendPoint, 7)
	byDate := map[string]*DashboardTrendPoint{}
	for i := 0; i < 7; i++ {
		day := today.AddDate(0, 0, i-6)
		key := day.Format("2006-01-02")
		points[i] = DashboardTrendPoint{Date: day.Format("1/2")}
		byDate[key] = &points[i]
	}

	var conns []model.AuditConn
	// ponytail: Go-side seven-day grouping stays DB-portable; switch to SQL GROUP BY if audit_conn grows large.
	service.DB.Select("created_at", "peer_id", "from_peer").
		Where("created_at >= ?", weekStart).
		Find(&conns)

	devicesByDate := map[string]map[string]bool{}
	for _, conn := range conns {
		createdAt := time.Time(conn.CreatedAt)
		key := createdAt.Format("2006-01-02")
		point := byDate[key]
		if point == nil {
			continue
		}
		point.Connections++
		if devicesByDate[key] == nil {
			devicesByDate[key] = map[string]bool{}
		}
		if conn.PeerId != "" {
			devicesByDate[key][conn.PeerId] = true
		}
		if conn.FromPeer != "" {
			devicesByDate[key][conn.FromPeer] = true
		}
		point.Devices = len(devicesByDate[key])
	}
	return points
}

func (d *Dashboard) deviceDistribution(stats DashboardStats) []DashboardDistribution {
	total := stats.TotalDevices
	return []DashboardDistribution{
		{Label: "在线设备", Value: stats.OnlineDevices, Percent: percent(stats.OnlineDevices, total), Color: "#22c55e"},
		{Label: "离线设备", Value: stats.OfflineDevices, Percent: percent(stats.OfflineDevices, total), Color: "#78869a"},
		{Label: "告警设备", Value: stats.WarningDevices, Percent: percent(stats.WarningDevices, total), Color: "#f97316"},
		{Label: "未分组设备", Value: stats.UnassignedDevices, Percent: percent(stats.UnassignedDevices, total), Color: "#b8c3d2"},
	}
}

func (d *Dashboard) loginLogs() []DashboardLoginLog {
	var logs []model.LoginLog
	service.DB.Order("id desc").Limit(5).Find(&logs)
	users := userNames(logs)
	res := make([]DashboardLoginLog, 0, len(logs))
	for _, log := range logs {
		res = append(res, DashboardLoginLog{
			Time:   time.Time(log.CreatedAt).Format("2006-01-02 15:04:05"),
			User:   fallback(users[log.UserId], fmt.Sprintf("用户 %d", log.UserId)),
			IP:     fallback(log.Ip, "-"),
			Device: fallback(log.Platform, "-"),
			Result: "登录成功",
		})
	}
	return res
}

func (d *Dashboard) devices(onlineAfter int64) []DashboardDevice {
	var peers []model.Peer
	service.DB.Order("last_online_time desc").Limit(5).Find(&peers)
	res := make([]DashboardDevice, 0, len(peers))
	for _, peer := range peers {
		status := "离线"
		if peer.LastOnlineTime >= onlineAfter {
			status = "在线"
		}
		res = append(res, DashboardDevice{
			OS:             osKey(peer.Os),
			Name:           fallback(peer.Alias, fallback(peer.Hostname, peer.Id)),
			ID:             peer.Id,
			Status:         status,
			Group:          fmt.Sprintf("分组 %d", peer.GroupId),
			LastOnlineTime: peer.LastOnlineTime,
		})
	}
	return res
}

func (d *Dashboard) services() []DashboardService {
	dbStatus := "正常"
	if sqlDB, err := service.DB.DB(); err != nil || sqlDB.Ping() != nil {
		dbStatus = "异常"
	}
	return []DashboardService{
		{Name: "API 服务", Status: "正常"},
		{Name: "信令服务", Status: "正常"},
		{Name: "中继服务", Status: "正常"},
		{Name: "数据库", Status: dbStatus},
		{Name: "存储服务", Status: "正常"},
	}
}

func (d *Dashboard) uptime() string {
	start, err := time.ParseInLocation("2006-01-02 15:04:05", (&service.AppService{}).GetStartTime(), time.Local)
	if err != nil {
		return "-"
	}
	duration := time.Since(start)
	days := int(duration.Hours()) / 24
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60
	return fmt.Sprintf("%d 天 %d 小时 %d 分钟", days, hours, minutes)
}

func userNames(logs []model.LoginLog) map[uint]string {
	ids := make([]uint, 0, len(logs))
	seen := map[uint]bool{}
	for _, log := range logs {
		if log.UserId > 0 && !seen[log.UserId] {
			ids = append(ids, log.UserId)
			seen[log.UserId] = true
		}
	}
	if len(ids) == 0 {
		return map[uint]string{}
	}
	var users []model.User
	service.DB.Where("id in ?", ids).Find(&users)
	res := map[uint]string{}
	for _, user := range users {
		res[user.Id] = user.Username
	}
	return res
}

func percent(value, total int64) string {
	if total == 0 {
		return "0%"
	}
	return fmt.Sprintf("%.1f%%", float64(value)*100/float64(total))
}

func fallback(value, fallbackValue string) string {
	if value == "" {
		return fallbackValue
	}
	return value
}

func osKey(os string) string {
	value := strings.ToLower(os)
	switch {
	case strings.Contains(value, "mac"):
		return "mac"
	case strings.Contains(value, "linux"):
		return "linux"
	default:
		return "win"
	}
}
