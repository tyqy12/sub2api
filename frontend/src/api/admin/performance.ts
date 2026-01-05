/**
 * Performance Monitoring API endpoints
 */

import { apiClient } from '../client'

export interface ConcurrencyStatus {
  platform: string
  current_in_use: number
  max_capacity: number
  load_percentage: number
  active_requests: number
  waiting_in_queue: number
}

export interface ConcurrencyByGroup {
  group_id: number
  group_name: string
  platform: string
  current_in_use: number
  max_capacity: number
  load_percentage: number
  active_requests: number
  waiting_in_queue: number
}

export interface PerformanceMetrics {
  total_requests: number
  active_requests: number
  completed_requests: number
  failed_requests: number
  timeout_requests: number
  response_time: {
    avg_ms: number
    min_ms: number
    max_ms: number
    p50_ms: number
    p95_ms: number
    p99_ms: number
  }
  upstream_time: {
    avg_ms: number
    max_ms: number
  }
  wait_time: {
    avg_ms: number
    max_ms: number
  }
  slot_usage: {
    user_slot: number
    account_slot: number
    wait_queue: number
  }
  concurrency_by_platform: ConcurrencyStatus[]
  concurrency_by_group: ConcurrencyByGroup[]
  throughput: {
    requests_per_second: number
  }
  uptime: {
    start_time: string
    last_updated: string
  }
}

export interface SlowRequests {
  slow_requests: {
    avg_response_time_ms: number
    max_response_time_ms: number
    p99_response_time_ms: number
    avg_upstream_time_ms: number
    max_upstream_time_ms: number
    avg_wait_time_ms: number
    max_wait_time_ms: number
  }
}

export interface HealthCheck {
  status: string
  healthy: boolean
  checks: {
    active_requests: {
      value: number
      status: string
    }
    response_time: {
      p95_ms: number
      status: string
    }
    timeout_rate: {
      count: number
      status: string
    }
  }
  warnings: string[]
  metrics: {
    total_requests: number
    completed_requests: number
    failed_requests: number
    timeout_requests: number
    requests_per_second: number
  }
}

/**
 * 获取性能监控启用状态
 */
export async function getEnabled(): Promise<{ enabled: boolean }> {
  const { data } = await apiClient.get<{ enabled: boolean }>('/admin/performance/enabled')
  return data
}

/**
 * 设置性能监控启用状态
 */
export async function setEnabled(enabled: boolean): Promise<{ message: string }> {
  const { data } = await apiClient.post<{ message: string }>('/admin/performance/enable', { enabled })
  return data
}

/**
 * Get performance metrics
 */
export async function getMetrics(): Promise<PerformanceMetrics> {
  const { data } = await apiClient.get<PerformanceMetrics>('/admin/performance/metrics')
  return data
}

/**
 * Get slow request statistics
 */
export async function getSlowRequests(): Promise<SlowRequests> {
  const { data } = await apiClient.get<SlowRequests>('/admin/performance/slow')
  return data
}

/**
 * Health check with warnings
 */
export async function healthCheck(): Promise<HealthCheck> {
  const { data } = await apiClient.get<HealthCheck>('/admin/performance/health')
  return data
}

/**
 * Reset performance metrics
 */
export async function resetMetrics(): Promise<{ message: string }> {
  const { data } = await apiClient.post<{ message: string }>('/admin/performance/reset')
  return data
}

export const performanceAPI = {
  getMetrics,
  getSlowRequests,
  healthCheck,
  resetMetrics,
  getEnabled,
  setEnabled
}

export default performanceAPI
