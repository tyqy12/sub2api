package service

import (
	"sync"
	"sync/atomic"
	"time"
)

// PerformanceMetrics 性能指标结构
type PerformanceMetrics struct {
	// 请求统计
	TotalRequests     int64 `json:"total_requests"`
	ActiveRequests    int64 `json:"active_requests"`
	CompletedRequests int64 `json:"completed_requests"`
	FailedRequests    int64 `json:"failed_requests"`
	TimeoutRequests   int64 `json:"timeout_requests"`

	// 响应时间统计（毫秒）
	AvgResponseTimeMs float64 `json:"avg_response_time_ms"`
	MinResponseTimeMs int64   `json:"min_response_time_ms"`
	MaxResponseTimeMs int64   `json:"max_response_time_ms"`
	P50ResponseTimeMs int64   `json:"p50_response_time_ms"`
	P95ResponseTimeMs int64   `json:"p95_response_time_ms"`
	P99ResponseTimeMs int64   `json:"p99_response_time_ms"`

	// 上游 API 响应时间（毫秒）
	AvgUpstreamTimeMs float64 `json:"avg_upstream_time_ms"`
	MaxUpstreamTimeMs int64   `json:"max_upstream_time_ms"`

	// 等待时间统计（毫秒）
	AvgWaitTimeMs float64 `json:"avg_wait_time_ms"`
	MaxWaitTimeMs int64   `json:"max_wait_time_ms"`

	// 并发槽位使用
	UserSlotUsage         int64                `json:"user_slot_usage"`
	AccountSlotUsage      int64                `json:"account_slot_usage"`
	WaitQueueSize         int64                `json:"wait_queue_size"`
	ConcurrencyByPlatform []ConcurrencyStatus  `json:"concurrency_by_platform"`
	ConcurrencyByGroup    []ConcurrencyByGroup `json:"concurrency_by_group"`

	// 吞吐量
	RequestsPerSecond float64 `json:"requests_per_second"`
	TokensPerSecond   float64 `json:"tokens_per_second"`

	// 统计周期
	StartTime   string `json:"start_time"`
	LastUpdated string `json:"last_updated"`
}

// ConcurrencyStatus 按平台的并发状态
type ConcurrencyStatus struct {
	Platform       string  `json:"platform"` // anthropic, openai, gemini, antigravity
	CurrentInUse   int64   `json:"current_in_use"`
	MaxCapacity    int64   `json:"max_capacity"`
	LoadPercentage float64 `json:"load_percentage"`
	ActiveRequests int64   `json:"active_requests"`
	WaitingInQueue int64   `json:"waiting_in_queue"`
}

// ConcurrencyByGroup 按分组的并发状态
type ConcurrencyByGroup struct {
	GroupID        int64   `json:"group_id"`
	GroupName      string  `json:"group_name"`
	Platform       string  `json:"platform"`
	CurrentInUse   int64   `json:"current_in_use"`
	MaxCapacity    int64   `json:"max_capacity"`
	LoadPercentage float64 `json:"load_percentage"`
	ActiveRequests int64   `json:"active_requests"`
	WaitingInQueue int64   `json:"waiting_in_queue"`
}

// RequestRecord 请求记录用于计算延迟分布
type RequestRecord struct {
	StartTime    time.Time
	EndTime      time.Time
	UpstreamTime time.Duration
	WaitTime     time.Duration
	IsTimeout    bool
	IsFailed     bool
}

// PerformanceMonitor 性能监控服务
type PerformanceMonitor struct {
	mu sync.RWMutex

	// 是否启用性能监控
	enabled atomic.Bool

	// 基础统计
	totalRequests int64
	completedReqs int64
	failedReqs    int64
	timeoutReqs   int64

	// 活跃请求
	activeRequests int64

	// 时间统计（使用 atomic 更新）
	responseTimes   []int64 // 响应时间历史（用于计算分位数）
	upstreamTimes   []int64 // 上游响应时间历史
	waitTimes       []int64 // 等待时间历史
	maxResponseTime int64
	maxUpstreamTime int64
	maxWaitTime     int64

	// 滑动窗口（最近5分钟）
	windowStart    time.Time
	windowReqCount int64

	// 并发槽位使用
	userSlotUsage    int64
	accountSlotUsage int64
	waitQueueSize    int64

	// 按平台的并发状态
	concurrencyByPlatform map[string]*ConcurrencyStatus

	// 按分组的并发状态（groupID -> ConcurrencyByGroup）
	concurrencyByGroup map[int64]*ConcurrencyByGroup

	// 统计锁
	statsMu sync.RWMutex

	startTime time.Time
}

// NewPerformanceMonitor 创建性能监控服务（用于 Wire 注入）
func NewPerformanceMonitor() *PerformanceMonitor {
	monitor := &PerformanceMonitor{
		startTime:             time.Now(),
		windowStart:           time.Now(),
		concurrencyByPlatform: make(map[string]*ConcurrencyStatus),
		concurrencyByGroup:    make(map[int64]*ConcurrencyByGroup),
	}
	// 默认关闭性能监控
	monitor.enabled.Store(false)
	return monitor
}

// SetEnabled 设置是否启用性能监控
func (m *PerformanceMonitor) SetEnabled(enabled bool) {
	m.enabled.Store(enabled)
}

// IsEnabled 检查性能监控是否启用
func (m *PerformanceMonitor) IsEnabled() bool {
	return m.enabled.Load()
}

// RecordRequestStart 记录请求开始
func (m *PerformanceMonitor) RecordRequestStart() {
	atomic.AddInt64(&m.activeRequests, 1)
	atomic.AddInt64(&m.totalRequests, 1)
}

// RecordRequestEnd 记录请求结束
func (m *PerformanceMonitor) RecordRequestEnd(record RequestRecord) {
	atomic.AddInt64(&m.activeRequests, -1)
	atomic.AddInt64(&m.completedReqs, 1)

	responseTime := record.EndTime.Sub(record.StartTime).Milliseconds()
	m.updateResponseTime(responseTime)

	if record.UpstreamTime > 0 {
		m.updateUpstreamTime(record.UpstreamTime.Milliseconds())
	}

	if record.WaitTime > 0 {
		m.updateWaitTime(record.WaitTime.Milliseconds())
	}

	if record.IsTimeout {
		atomic.AddInt64(&m.timeoutReqs, 1)
	}
	if record.IsFailed {
		atomic.AddInt64(&m.failedReqs, 1)
	}

	// 更新滑动窗口
	m.statsMu.Lock()
	m.windowReqCount++
	m.statsMu.Unlock()
}

// RecordRequestComplete 记录请求完成（简化版，用于网关handler）
func (m *PerformanceMonitor) RecordRequestComplete(durationMs, upstreamDurationMs int64) {
	atomic.AddInt64(&m.activeRequests, -1)
	atomic.AddInt64(&m.completedReqs, 1)

	m.updateResponseTime(durationMs)

	if upstreamDurationMs > 0 {
		m.updateUpstreamTime(upstreamDurationMs)
	}

	// 更新滑动窗口
	m.statsMu.Lock()
	m.windowReqCount++
	m.statsMu.Unlock()
}

// updateResponseTime 更新响应时间统计
func (m *PerformanceMonitor) updateResponseTime(ms int64) {
	m.statsMu.Lock()
	defer m.statsMu.Unlock()

	m.responseTimes = append(m.responseTimes, ms)
	if ms > m.maxResponseTime {
		m.maxResponseTime = ms
	}

	// 保持最近 1000 条记录
	if len(m.responseTimes) > 1000 {
		m.responseTimes = m.responseTimes[len(m.responseTimes)-1000:]
	}
}

// updateUpstreamTime 更新上游响应时间
func (m *PerformanceMonitor) updateUpstreamTime(ms int64) {
	m.statsMu.Lock()
	defer m.statsMu.Unlock()

	m.upstreamTimes = append(m.upstreamTimes, ms)
	if ms > m.maxUpstreamTime {
		m.maxUpstreamTime = ms
	}

	if len(m.upstreamTimes) > 1000 {
		m.upstreamTimes = m.upstreamTimes[len(m.upstreamTimes)-1000:]
	}
}

// updateWaitTime 更新等待时间
func (m *PerformanceMonitor) updateWaitTime(ms int64) {
	m.statsMu.Lock()
	defer m.statsMu.Unlock()

	m.waitTimes = append(m.waitTimes, ms)
	if ms > m.maxWaitTime {
		m.maxWaitTime = ms
	}

	if len(m.waitTimes) > 1000 {
		m.waitTimes = m.waitTimes[len(m.waitTimes)-1000:]
	}
}

// UpdateSlotUsage 更新槽位使用情况
func (m *PerformanceMonitor) UpdateSlotUsage(user, account, wait int64) {
	atomic.StoreInt64(&m.userSlotUsage, user)
	atomic.StoreInt64(&m.accountSlotUsage, account)
	atomic.StoreInt64(&m.waitQueueSize, wait)
}

// UpdateConcurrencyByPlatform 更新按平台的并发状态（接收完整map）
func (m *PerformanceMonitor) UpdateConcurrencyByPlatform(concurrencyData map[string]map[string]int64) {
	m.statsMu.Lock()
	defer m.statsMu.Unlock()

	for platform, data := range concurrencyData {
		status, exists := m.concurrencyByPlatform[platform]
		if !exists {
			status = &ConcurrencyStatus{
				Platform: platform,
			}
			m.concurrencyByPlatform[platform] = status
		}

		status.CurrentInUse = data["current"]
		status.MaxCapacity = data["max"]
		status.WaitingInQueue = data["waiting"]
		status.ActiveRequests = data["active"]

		if status.MaxCapacity > 0 {
			status.LoadPercentage = float64(status.CurrentInUse) / float64(status.MaxCapacity) * 100
		} else {
			status.LoadPercentage = 0
		}
	}
}

// UpdateConcurrencyByGroup 更新按分组的并发状态
// concurrencyData: map[groupID]map[string]interface{}{"current":, "max":, "waiting":, "group_name":, "platform":}
func (m *PerformanceMonitor) UpdateConcurrencyByGroup(concurrencyData map[int64]map[string]interface{}) {
	m.statsMu.Lock()
	defer m.statsMu.Unlock()

	for groupID, data := range concurrencyData {
		status, exists := m.concurrencyByGroup[groupID]
		if !exists {
			status = &ConcurrencyByGroup{
				GroupID: groupID,
			}
			m.concurrencyByGroup[groupID] = status
		}

		if groupName, ok := data["group_name"].(string); ok {
			status.GroupName = groupName
		}
		if platform, ok := data["platform"].(string); ok {
			status.Platform = platform
		}
		if current, ok := data["current"].(int64); ok {
			status.CurrentInUse = current
		}
		if maxCap, ok := data["max"].(int64); ok {
			status.MaxCapacity = maxCap
		}
		if waiting, ok := data["waiting"].(int64); ok {
			status.WaitingInQueue = waiting
		}

		if status.MaxCapacity > 0 {
			status.LoadPercentage = float64(status.CurrentInUse) / float64(status.MaxCapacity) * 100
		} else {
			status.LoadPercentage = 0
		}
	}
}

// GetMetrics 获取当前性能指标
func (m *PerformanceMonitor) GetMetrics() PerformanceMetrics {
	m.statsMu.RLock()
	defer m.statsMu.RUnlock()

	// 计算平均响应时间
	var totalResponseTime int64
	for _, t := range m.responseTimes {
		totalResponseTime += t
	}
	avgResponseTime := float64(0)
	if len(m.responseTimes) > 0 {
		avgResponseTime = float64(totalResponseTime) / float64(len(m.responseTimes))
	}

	// 计算平均上游时间
	var totalUpstreamTime int64
	for _, t := range m.upstreamTimes {
		totalUpstreamTime += t
	}
	avgUpstreamTime := float64(0)
	if len(m.upstreamTimes) > 0 {
		avgUpstreamTime = float64(totalUpstreamTime) / float64(len(m.upstreamTimes))
	}

	// 计算平均等待时间
	var totalWaitTime int64
	for _, t := range m.waitTimes {
		totalWaitTime += t
	}
	avgWaitTime := float64(0)
	if len(m.waitTimes) > 0 {
		avgWaitTime = float64(totalWaitTime) / float64(len(m.waitTimes))
	}

	// 计算分位数
	p50 := percentile(m.responseTimes, 50)
	p95 := percentile(m.responseTimes, 95)
	p99 := percentile(m.responseTimes, 99)

	// 计算 RPS
	windowDuration := time.Since(m.windowStart)
	rps := float64(0)
	if windowDuration > 0 {
		rps = float64(m.windowReqCount) / windowDuration.Seconds()
	}

	// 收集按平台的并发状态
	var concurrencyByPlatformList []ConcurrencyStatus
	for _, status := range m.concurrencyByPlatform {
		concurrencyByPlatformList = append(concurrencyByPlatformList, *status)
	}

	// 收集按分组的并发状态
	var concurrencyByGroupList []ConcurrencyByGroup
	for _, status := range m.concurrencyByGroup {
		concurrencyByGroupList = append(concurrencyByGroupList, *status)
	}

	return PerformanceMetrics{
		TotalRequests:         atomic.LoadInt64(&m.totalRequests),
		ActiveRequests:        atomic.LoadInt64(&m.activeRequests),
		CompletedRequests:     atomic.LoadInt64(&m.completedReqs),
		FailedRequests:        atomic.LoadInt64(&m.failedReqs),
		TimeoutRequests:       atomic.LoadInt64(&m.timeoutReqs),
		AvgResponseTimeMs:     avgResponseTime,
		MinResponseTimeMs:     m.minTime(m.responseTimes),
		MaxResponseTimeMs:     m.maxResponseTime,
		P50ResponseTimeMs:     p50,
		P95ResponseTimeMs:     p95,
		P99ResponseTimeMs:     p99,
		AvgUpstreamTimeMs:     avgUpstreamTime,
		MaxUpstreamTimeMs:     m.maxUpstreamTime,
		AvgWaitTimeMs:         avgWaitTime,
		MaxWaitTimeMs:         m.maxWaitTime,
		UserSlotUsage:         atomic.LoadInt64(&m.userSlotUsage),
		AccountSlotUsage:      atomic.LoadInt64(&m.accountSlotUsage),
		WaitQueueSize:         atomic.LoadInt64(&m.waitQueueSize),
		ConcurrencyByPlatform: concurrencyByPlatformList,
		ConcurrencyByGroup:    concurrencyByGroupList,
		RequestsPerSecond:     rps,
		StartTime:             m.startTime.Format(time.RFC3339),
		LastUpdated:           time.Now().Format(time.RFC3339),
	}
}

// minTime 计算最小值
func (m *PerformanceMonitor) minTime(times []int64) int64 {
	if len(times) == 0 {
		return 0
	}
	min := times[0]
	for _, t := range times {
		if t < min {
			min = t
		}
	}
	return min
}

// percentile 计算分位数
func percentile(times []int64, p int) int64 {
	if len(times) == 0 {
		return 0
	}

	// 复制并排序
	sorted := make([]int64, len(times))
	copy(sorted, times)
	quickselect(sorted, 0, len(sorted)-1, len(sorted)*p/100)

	index := len(sorted) * p / 100
	if index >= len(sorted) {
		index = len(sorted) - 1
	}
	return sorted[index]
}

// quickselect 快速选择算法
func quickselect(arr []int64, left, right, k int) {
	if left == right {
		return
	}
	pivotIndex := partition(arr, left, right)
	if k == pivotIndex {
		return
	} else if k < pivotIndex {
		quickselect(arr, left, pivotIndex-1, k)
	} else {
		quickselect(arr, pivotIndex+1, right, k)
	}
}

// partition 分区函数
func partition(arr []int64, left, right int) int {
	pivot := arr[right]
	i := left - 1
	for j := left; j < right; j++ {
		if arr[j] <= pivot {
			i++
			arr[i], arr[j] = arr[j], arr[i]
		}
	}
	arr[i+1], arr[right] = arr[right], arr[i+1]
	return i + 1
}

// Reset 重置统计
func (m *PerformanceMonitor) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	atomic.StoreInt64(&m.totalRequests, 0)
	atomic.StoreInt64(&m.completedReqs, 0)
	atomic.StoreInt64(&m.failedReqs, 0)
	atomic.StoreInt64(&m.timeoutReqs, 0)
	atomic.StoreInt64(&m.activeRequests, 0)
	atomic.StoreInt64(&m.userSlotUsage, 0)
	atomic.StoreInt64(&m.accountSlotUsage, 0)
	atomic.StoreInt64(&m.waitQueueSize, 0)

	m.statsMu.Lock()
	m.responseTimes = nil
	m.upstreamTimes = nil
	m.waitTimes = nil
	m.maxResponseTime = 0
	m.maxUpstreamTime = 0
	m.maxWaitTime = 0
	m.windowReqCount = 0
	m.concurrencyByPlatform = make(map[string]*ConcurrencyStatus)
	m.concurrencyByGroup = make(map[int64]*ConcurrencyByGroup)
	m.statsMu.Unlock()

	m.startTime = time.Now()
	m.windowStart = time.Now()
}
