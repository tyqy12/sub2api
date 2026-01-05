package admin

import (
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// PerformanceHandler handles performance monitoring
type PerformanceHandler struct {
	monitor *service.PerformanceMonitor
}

// NewPerformanceHandler creates a new performance handler
func NewPerformanceHandler(monitor *service.PerformanceMonitor) *PerformanceHandler {
	return &PerformanceHandler{
		monitor: monitor,
	}
}

// GetMetrics handles getting performance metrics
// GET /api/v1/admin/performance/metrics
func (h *PerformanceHandler) GetMetrics(c *gin.Context) {
	// 检查监控是否启用
	if !h.monitor.IsEnabled() {
		response.Success(c, gin.H{
			"enabled": false,
			"message": "性能监控已关闭，请在右上角开启",
		})
		return
	}
	metrics := h.monitor.GetMetrics()
	response.Success(c, gin.H{
		"enabled":            true,
		"total_requests":     metrics.TotalRequests,
		"active_requests":    metrics.ActiveRequests,
		"completed_requests": metrics.CompletedRequests,
		"failed_requests":    metrics.FailedRequests,
		"timeout_requests":   metrics.TimeoutRequests,
		"response_time": gin.H{
			"avg_ms": metrics.AvgResponseTimeMs,
			"min_ms": metrics.MinResponseTimeMs,
			"max_ms": metrics.MaxResponseTimeMs,
			"p50_ms": metrics.P50ResponseTimeMs,
			"p95_ms": metrics.P95ResponseTimeMs,
			"p99_ms": metrics.P99ResponseTimeMs,
		},
		"upstream_time": gin.H{
			"avg_ms": metrics.AvgUpstreamTimeMs,
			"max_ms": metrics.MaxUpstreamTimeMs,
		},
		"wait_time": gin.H{
			"avg_ms": metrics.AvgWaitTimeMs,
			"max_ms": metrics.MaxWaitTimeMs,
		},
		"slot_usage": gin.H{
			"user_slot":    metrics.UserSlotUsage,
			"account_slot": metrics.AccountSlotUsage,
			"wait_queue":   metrics.WaitQueueSize,
		},
		"concurrency_by_platform": metrics.ConcurrencyByPlatform,
		"concurrency_by_group":    metrics.ConcurrencyByGroup,
		"throughput": gin.H{
			"requests_per_second": metrics.RequestsPerSecond,
		},
		"uptime": gin.H{
			"start_time":   metrics.StartTime,
			"last_updated": metrics.LastUpdated,
		},
	})
}

// ResetMetrics handles resetting performance metrics
// POST /api/v1/admin/performance/reset
func (h *PerformanceHandler) ResetMetrics(c *gin.Context) {
	h.monitor.Reset()
	response.Success(c, gin.H{"message": "Metrics reset successfully"})
}

// SetEnabled handles enabling/disabling performance monitoring
// POST /api/v1/admin/performance/enable
func (h *PerformanceHandler) SetEnabled(c *gin.Context) {
	var req struct {
		Enabled *bool `json:"enabled" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "Invalid request body")
		return
	}
	h.monitor.SetEnabled(*req.Enabled)
	if *req.Enabled {
		response.Success(c, gin.H{"message": "性能监控已开启"})
	} else {
		response.Success(c, gin.H{"message": "性能监控已关闭"})
	}
}

// GetEnabled handles getting the enabled status
// GET /api/v1/admin/performance/enabled
func (h *PerformanceHandler) GetEnabled(c *gin.Context) {
	response.Success(c, gin.H{
		"enabled": h.monitor.IsEnabled(),
	})
}

// GetSlowRequests handles getting slow request statistics
// GET /api/v1/admin/performance/slow
func (h *PerformanceHandler) GetSlowRequests(c *gin.Context) {
	metrics := h.monitor.GetMetrics()
	response.Success(c, gin.H{
		"slow_requests": gin.H{
			"avg_response_time_ms": metrics.AvgResponseTimeMs,
			"max_response_time_ms": metrics.MaxResponseTimeMs,
			"p99_response_time_ms": metrics.P99ResponseTimeMs,
			"avg_upstream_time_ms": metrics.AvgUpstreamTimeMs,
			"max_upstream_time_ms": metrics.MaxUpstreamTimeMs,
			"avg_wait_time_ms":     metrics.AvgWaitTimeMs,
			"max_wait_time_ms":     metrics.MaxWaitTimeMs,
		},
	})
}

// HealthCheck performs a health check
// GET /api/v1/admin/performance/health
func (h *PerformanceHandler) HealthCheck(c *gin.Context) {
	metrics := h.monitor.GetMetrics()

	// 健康状态评估
	status := "healthy"
	warnings := []string{}

	// 检查响应时间
	if metrics.P95ResponseTimeMs > 60000 {
		status = "degraded"
		warnings = append(warnings, "95%响应时间超过60秒")
	}

	// 检查超时率
	total := metrics.CompletedRequests + metrics.FailedRequests + metrics.TimeoutRequests
	if total > 0 {
		timeoutRate := float64(metrics.TimeoutRequests) / float64(total) * 100
		if timeoutRate > 5 {
			status = "unhealthy"
			warnings = append(warnings, "超时率超过5%")
		} else if timeoutRate > 1 {
			status = "degraded"
			warnings = append(warnings, "超时率超过1%")
		}
	}

	// 检查活跃请求数
	if metrics.ActiveRequests > 1000 {
		status = "degraded"
		warnings = append(warnings, "活跃请求数超过1000")
	}

	response.Success(c, gin.H{
		"status":  status,
		"healthy": status == "healthy",
		"checks": gin.H{
			"active_requests": gin.H{
				"value":  metrics.ActiveRequests,
				"status": "ok",
			},
			"response_time": gin.H{
				"p95_ms": metrics.P95ResponseTimeMs,
				"status": "ok",
			},
			"timeout_rate": gin.H{
				"count":  metrics.TimeoutRequests,
				"status": "ok",
			},
		},
		"warnings": warnings,
		"metrics": gin.H{
			"total_requests":      metrics.TotalRequests,
			"completed_requests":  metrics.CompletedRequests,
			"failed_requests":     metrics.FailedRequests,
			"timeout_requests":    metrics.TimeoutRequests,
			"requests_per_second": metrics.RequestsPerSecond,
		},
	})
}
