<template>
  <AppLayout>
    <div class="space-y-6">
      <!-- Header with Health Status -->
      <div class="flex flex-wrap items-center justify-between gap-4">
        <div class="flex items-center gap-3">
          <!-- Enable/Disable Toggle -->
          <div class="flex items-center gap-2">
            <label class="relative inline-flex items-center cursor-pointer">
              <input
                type="checkbox"
                :checked="enabled"
                @click="toggleEnabled"
                class="sr-only peer"
                :disabled="toggling"
              >
              <div class="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 dark:peer-focus:ring-blue-800 rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full rtl:peer-checked:after:-translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:start-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-gray-600 peer-checked:bg-blue-600"></div>
              <span class="ms-3 text-sm font-medium text-gray-900 dark:text-white">
                {{ enabled ? t('common.enabled') : t('common.disabled') }}
              </span>
            </label>
          </div>
          <!-- Health Badge -->
          <div
            v-if="healthData && enabled"
            class="flex items-center gap-2 rounded-full px-3 py-1 text-sm font-medium"
            :class="{
              'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400': healthData.healthy,
              'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400': !healthData.healthy && healthData.status === 'degraded',
              'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400': healthData.status === 'unhealthy'
            }"
          >
            <span
              class="h-2 w-2 rounded-full animate-pulse"
              :class="{
                'bg-green-500': healthData.healthy,
                'bg-yellow-500': healthData.status === 'degraded',
                'bg-red-500': healthData.status === 'unhealthy'
              }"
            ></span>
            {{ t(`admin.performance.health.${healthData.status}`) }}
          </div>
        </div>
        <!-- Action Buttons -->
        <div class="flex items-center gap-2">
          <!-- Refresh Button -->
          <button
            v-if="enabled"
            @click="refreshData"
            class="inline-flex items-center gap-2 px-4 py-2 bg-blue-50 hover:bg-blue-100 dark:bg-blue-900/30 dark:hover:bg-blue-900/50 text-blue-700 dark:text-blue-300 rounded-lg transition-colors duration-200 disabled:opacity-50"
            :disabled="loading"
          >
            <svg 
              class="h-4 w-4" 
              :class="{ 'animate-spin': loading }"
              fill="none" 
              stroke="currentColor" 
              viewBox="0 0 24 24"
            >
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
            </svg>
            <span class="text-sm font-medium">{{ t('common.refresh') }}</span>
          </button>
          <!-- Reset Button -->
          <button
            v-if="enabled"
            @click="resetMetrics"
            class="inline-flex items-center gap-2 px-4 py-2 bg-red-50 hover:bg-red-100 dark:bg-red-900/30 dark:hover:bg-red-900/50 text-red-700 dark:text-red-300 rounded-lg transition-colors duration-200 disabled:opacity-50"
            :disabled="resetting"
          >
            <svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
            </svg>
            <span class="text-sm font-medium">{{ resetting ? t('common.loading') : t('admin.performance.reset') }}</span>
          </button>
        </div>
      </div>

      <!-- Disabled State -->
      <div v-if="!enabled" class="card p-8 text-center">
        <div class="flex flex-col items-center">
          <div class="relative">
            <svg class="h-16 w-16 text-gray-400 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
            </svg>
            <div class="absolute -bottom-1 -right-1 bg-gray-200 dark:bg-gray-600 rounded-full p-1">
              <svg class="h-4 w-4 text-gray-500 dark:text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M18.364 18.364A9 9 0 005.636 5.636m12.728 12.728A9 9 0 015.636 5.636m12.728 12.728L5.636 5.636" />
              </svg>
            </div>
          </div>
          <h3 class="text-lg font-medium text-gray-900 dark:text-white mb-2">
            {{ t('admin.performance.disabledTitle') }}
          </h3>
          <p class="text-gray-500 dark:text-gray-400 mb-6 max-w-md">
            {{ t('admin.performance.disabledDesc') }}
          </p>
          <button
            @click="enableMonitoring"
            class="inline-flex items-center gap-2 px-6 py-3 bg-gradient-to-r from-blue-500 to-blue-600 hover:from-blue-600 hover:to-blue-700 text-white rounded-lg transition-all duration-200 shadow-md hover:shadow-lg disabled:opacity-50 disabled:cursor-not-allowed"
            :disabled="enabling"
          >
            <svg v-if="enabling" class="h-5 w-5 animate-spin" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
            <svg v-else class="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
            </svg>
            <span class="font-medium">{{ enabling ? t('common.loading') : t('admin.performance.enableButton') }}</span>
          </button>
        </div>
      </div>

      <!-- Warnings -->
      <div v-if="healthData?.warnings?.length && enabled" class="card border-l-4 border-yellow-500 p-4">
        <div class="flex">
          <div class="flex-shrink-0">
            <svg class="h-5 w-5 text-yellow-400" viewBox="0 0 20 20" fill="currentColor">
              <path fill-rule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clip-rule="evenodd" />
            </svg>
          </div>
          <div class="ml-3">
            <h3 class="text-sm font-medium text-yellow-800 dark:text-yellow-200">
              {{ t('admin.performance.warningTitle') }}
            </h3>
            <div class="mt-2 text-sm text-yellow-700 dark:text-yellow-300">
              <ul class="list-disc space-y-1 pl-5">
                <li v-for="(warning, idx) in healthData.warnings" :key="idx">
                  {{ warning }}
                </li>
              </ul>
            </div>
          </div>
        </div>
      </div>

      <!-- Loading State -->
      <div v-if="loading && enabled" class="flex items-center justify-center py-16">
        <div class="text-center">
          <LoadingSpinner />
          <p class="mt-4 text-sm text-gray-500 dark:text-gray-400">{{ t('common.loading') }}</p>
        </div>
      </div>

      <template v-else-if="metrics && enabled">
        <!-- Row 1: Request Stats -->
        <div class="grid grid-cols-2 gap-4 lg:grid-cols-5">
          <!-- Total Requests -->
          <div class="card p-4 hover:shadow-md transition-shadow">
            <div class="flex items-center gap-3">
              <div class="rounded-lg bg-blue-100 p-2.5 dark:bg-blue-900/30">
                <svg class="h-5 w-5 text-blue-600 dark:text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                </svg>
              </div>
              <div>
                <p class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.performance.requests.total') }}</p>
                <p class="text-xl font-bold text-gray-900 dark:text-white">{{ formatNumber(metrics.total_requests) }}</p>
              </div>
            </div>
          </div>

          <!-- Active Requests -->
          <div class="card p-4 hover:shadow-md transition-shadow">
            <div class="flex items-center gap-3">
              <div class="rounded-lg bg-green-100 p-2.5 dark:bg-green-900/30">
                <svg class="h-5 w-5 text-green-600 dark:text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
                </svg>
              </div>
              <div>
                <p class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.performance.requests.active') }}</p>
                <p class="text-xl font-bold text-gray-900 dark:text-white">{{ formatNumber(metrics.active_requests) }}</p>
              </div>
            </div>
          </div>

          <!-- Completed -->
          <div class="card p-4 hover:shadow-md transition-shadow">
            <div class="flex items-center gap-3">
              <div class="rounded-lg bg-emerald-100 p-2.5 dark:bg-emerald-900/30">
                <svg class="h-5 w-5 text-emerald-600 dark:text-emerald-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 0 0 1 8 0z" />
                </svg>
              </div>
              <div>
                <p class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.performance.requests.completed') }}</p>
                <p class="text-xl font-bold text-gray-900 dark:text-white">{{ formatNumber(metrics.completed_requests) }}</p>
              </div>
            </div>
          </div>

          <!-- Failed -->
          <div class="card p-4 hover:shadow-md transition-shadow">
            <div class="flex items-center gap-3">
              <div class="rounded-lg bg-red-100 p-2.5 dark:bg-red-900/30">
                <svg class="h-5 w-5 text-red-600 dark:text-red-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 0 0 1 8 0z" />
                </svg>
              </div>
              <div>
                <p class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.performance.requests.failed') }}</p>
                <p class="text-xl font-bold text-gray-900 dark:text-white">{{ formatNumber(metrics.failed_requests) }}</p>
              </div>
            </div>
          </div>

          <!-- Timeout -->
          <div class="card p-4 hover:shadow-md transition-shadow">
            <div class="flex items-center gap-3">
              <div class="rounded-lg bg-orange-100 p-2.5 dark:bg-orange-900/30">
                <svg class="h-5 w-5 text-orange-600 dark:text-orange-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 0 0 1 8 0z" />
                </svg>
              </div>
              <div>
                <p class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.performance.requests.timeout') }}</p>
                <p class="text-xl font-bold text-gray-900 dark:text-white">{{ formatNumber(metrics.timeout_requests) }}</p>
              </div>
            </div>
          </div>
        </div>

        <!-- Row 2: Response Time & Throughput -->
        <div class="grid grid-cols-2 gap-4 lg:grid-cols-5">
          <!-- Average Response -->
          <div class="card p-4 hover:shadow-md transition-shadow">
            <div class="flex items-center gap-3">
              <div class="rounded-lg bg-violet-100 p-2.5 dark:bg-violet-900/30">
                <svg class="h-5 w-5 text-violet-600 dark:text-violet-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 0 0 1 8 0z" />
                </svg>
              </div>
              <div>
                <p class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.performance.responseTime.average') }}</p>
                <p class="text-xl font-bold text-gray-900 dark:text-white">{{ formatDuration(metrics.response_time.avg_ms) }}</p>
              </div>
            </div>
          </div>

          <!-- Min Response -->
          <div class="card p-4 hover:shadow-md transition-shadow">
            <div class="flex items-center gap-3">
              <div class="rounded-lg bg-teal-100 p-2.5 dark:bg-teal-900/30">
                <svg class="h-5 w-5 text-teal-600 dark:text-teal-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 17h8m0 0V9m0 8l-8-8-4 4-6-6" />
                </svg>
              </div>
              <div>
                <p class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.performance.responseTime.min') }}</p>
                <p class="text-xl font-bold text-gray-900 dark:text-white">{{ formatDuration(metrics.response_time.min_ms) }}</p>
              </div>
            </div>
          </div>

          <!-- P50 Response -->
          <div class="card p-4 hover:shadow-md transition-shadow">
            <div class="flex items-center gap-3">
              <div class="rounded-lg bg-indigo-100 p-2.5 dark:bg-indigo-900/30">
                <svg class="h-5 w-5 text-indigo-600 dark:text-indigo-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
                </svg>
              </div>
              <div>
                <p class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.performance.responseTime.p50') }}</p>
                <p class="text-xl font-bold text-gray-900 dark:text-white">{{ formatDuration(metrics.response_time.p50_ms) }}</p>
              </div>
            </div>
          </div>

          <!-- P95 Response -->
          <div class="card p-4 hover:shadow-md transition-shadow">
            <div class="flex items-center gap-3">
              <div class="rounded-lg bg-amber-100 p-2.5 dark:bg-amber-900/30">
                <svg class="h-5 w-5 text-amber-600 dark:text-amber-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7h8m0 0v8m0-8l-8 8-4-4-6-6" />
                </svg>
              </div>
              <div>
                <p class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.performance.responseTime.p95') }}</p>
                <p class="text-xl font-bold text-gray-900 dark:text-white">{{ formatDuration(metrics.response_time.p95_ms) }}</p>
              </div>
            </div>
          </div>

          <!-- RPS -->
          <div class="card p-4 hover:shadow-md transition-shadow">
            <div class="flex items-center gap-3">
              <div class="rounded-lg bg-cyan-100 p-2.5 dark:bg-cyan-900/30">
                <svg class="h-5 w-5 text-cyan-600 dark:text-cyan-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7h8m0 0v8m0-8l-8 8-4-4-6-6" />
                </svg>
              </div>
              <div>
                <p class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.performance.throughput.rps') }}</p>
                <p class="text-xl font-bold text-gray-900 dark:text-white">{{ metrics.throughput.requests_per_second.toFixed(1) }}</p>
              </div>
            </div>
          </div>
        </div>

        <!-- Row 3: Upstream & Wait Time -->
        <div class="grid grid-cols-2 gap-4 lg:grid-cols-4">
          <!-- Upstream Avg -->
          <div class="card p-4 hover:shadow-md transition-shadow">
            <div class="flex items-center gap-3">
              <div class="rounded-lg bg-sky-100 p-2.5 dark:bg-sky-900/30">
                <svg class="h-5 w-5 text-sky-600 dark:text-sky-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3.75 13.5l10.5-11.25L12 10.5h8.25L9.75 21.75 12 13.5H3.75z" />
                </svg>
              </div>
              <div>
                <p class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.performance.upstream.avgResponse') }}</p>
                <p class="text-xl font-bold text-gray-900 dark:text-white">{{ formatDuration(metrics.upstream_time.avg_ms) }}</p>
              </div>
            </div>
          </div>

          <!-- Upstream Max -->
          <div class="card p-4 hover:shadow-md transition-shadow">
            <div class="flex items-center gap-3">
              <div class="rounded-lg bg-blue-100 p-2.5 dark:bg-blue-900/30">
                <svg class="h-5 w-5 text-blue-600 dark:text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 17h8m0 0V9m0 8l-8-8-4 4-6-6" />
                </svg>
              </div>
              <div>
                <p class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.performance.upstream.maxUpstream') }}</p>
                <p class="text-xl font-bold text-gray-900 dark:text-white">{{ formatDuration(metrics.upstream_time.max_ms) }}</p>
              </div>
            </div>
          </div>

          <!-- Wait Time Avg -->
          <div class="card p-4 hover:shadow-md transition-shadow">
            <div class="flex items-center gap-3">
              <div class="rounded-lg bg-yellow-100 p-2.5 dark:bg-yellow-900/30">
                <svg class="h-5 w-5 text-yellow-600 dark:text-yellow-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 0 0 1 8 0z" />
                </svg>
              </div>
              <div>
                <p class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.performance.upstream.avgWait') }}</p>
                <p class="text-xl font-bold text-gray-900 dark:text-white">{{ formatDuration(metrics.wait_time.avg_ms) }}</p>
              </div>
            </div>
          </div>

          <!-- Wait Time Max -->
          <div class="card p-4 hover:shadow-md transition-shadow">
            <div class="flex items-center gap-3">
              <div class="rounded-lg bg-orange-100 p-2.5 dark:bg-orange-900/30">
                <svg class="h-5 w-5 text-orange-600 dark:text-orange-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 0 0 1 8 0z" />
                </svg>
              </div>
              <div>
                <p class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.performance.maxWait') }}</p>
                <p class="text-xl font-bold text-gray-900 dark:text-white">{{ formatDuration(metrics.wait_time.max_ms) }}</p>
              </div>
            </div>
          </div>
        </div>

        <!-- Row 4: Concurrency by Platform -->
        <div class="card p-6">
          <div class="flex items-center justify-between mb-4">
            <h3 class="text-base font-semibold text-gray-900 dark:text-white">
              {{ t('admin.performance.concurrency.byPlatform') }}
            </h3>
          </div>
          <div v-if="metrics.concurrency_by_platform && metrics.concurrency_by_platform.length > 0" class="space-y-4">
            <div v-for="platform in metrics.concurrency_by_platform" :key="platform.platform" class="border-l-4 pl-4 rounded-r-lg"
              :class="{
                'border-orange-500 bg-orange-50 dark:bg-orange-900/10': platform.platform === 'anthropic',
                'border-blue-500 bg-blue-50 dark:bg-blue-900/10': platform.platform === 'openai',
                'border-purple-500 bg-purple-50 dark:bg-purple-900/10': platform.platform === 'gemini',
                'border-green-500 bg-green-50 dark:bg-green-900/10': platform.platform === 'antigravity'
              }">
              <div class="flex items-center justify-between mb-2">
                <div class="flex items-center gap-3">
                  <span class="text-sm font-semibold text-gray-900 dark:text-white">
                    {{ t(`admin.performance.concurrency.platforms.${platform.platform}`) }}
                  </span>
                  <span class="text-xs text-gray-500 dark:text-gray-400 bg-white dark:bg-gray-800 px-2 py-0.5 rounded">
                    ({{ formatNumber(platform.current_in_use) }} / {{ formatNumber(platform.max_capacity) }})
                  </span>
                </div>
                <span class="text-sm font-bold"
                  :class="{
                    'text-green-600 dark:text-green-400': platform.load_percentage < 50,
                    'text-yellow-600 dark:text-yellow-400': platform.load_percentage >= 50 && platform.load_percentage < 80,
                    'text-red-600 dark:text-red-400': platform.load_percentage >= 80
                  }">
                  {{ platform.load_percentage.toFixed(1) }}%
                </span>
              </div>
              <div class="h-3 overflow-hidden rounded-full bg-gray-200 dark:bg-gray-700">
                <div
                  class="h-full rounded-full transition-all duration-500"
                  :class="{
                    'bg-orange-500': platform.platform === 'anthropic',
                    'bg-blue-500': platform.platform === 'openai',
                    'bg-purple-500': platform.platform === 'gemini',
                    'bg-green-500': platform.platform === 'antigravity'
                  }"
                  :style="{ width: Math.min(platform.load_percentage, 100) + '%' }"></div>
              </div>
              <div class="mt-2 flex items-center justify-between text-xs text-gray-500 dark:text-gray-400">
                <span>{{ t('admin.performance.concurrency.activeRequests') }}: {{ formatNumber(platform.active_requests) }}</span>
                <span>{{ t('admin.performance.concurrency.waitingInQueue') }}: {{ formatNumber(platform.waiting_in_queue) }}</span>
              </div>
            </div>
          </div>
          <div v-else class="text-center py-8">
            <svg class="h-12 w-12 text-gray-400 mx-auto mb-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4" />
            </svg>
            <p class="text-sm text-gray-500 dark:text-gray-400">{{ t('common.noData') }}</p>
          </div>
        </div>

        <!-- Row 5: Concurrency by Group -->
        <div class="card p-6">
          <div class="flex items-center justify-between mb-4">
            <h3 class="text-base font-semibold text-gray-900 dark:text-white">
              {{ t('admin.performance.concurrency.byGroup') }}
            </h3>
          </div>
          <div v-if="metrics.concurrency_by_group && metrics.concurrency_by_group.length > 0" class="space-y-4">
            <div v-for="group in metrics.concurrency_by_group" :key="`${group.group_id}-${group.platform}`" class="border-l-4 pl-4 rounded-r-lg"
              :class="{
                'border-orange-500 bg-orange-50 dark:bg-orange-900/10': group.platform === 'anthropic',
                'border-blue-500 bg-blue-50 dark:bg-blue-900/10': group.platform === 'openai',
                'border-purple-500 bg-purple-50 dark:bg-purple-900/10': group.platform === 'gemini',
                'border-green-500 bg-green-50 dark:bg-green-900/10': group.platform === 'antigravity'
              }">
              <div class="flex items-center justify-between mb-2">
                <div class="flex items-center gap-3 flex-wrap">
                  <span class="text-sm font-semibold text-gray-900 dark:text-white">{{ group.group_name }}</span>
                  <span class="rounded px-2 py-0.5 text-xs font-medium"
                    :class="{
                      'bg-orange-100 text-orange-700 dark:bg-orange-900/50 dark:text-orange-300': group.platform === 'anthropic',
                      'bg-blue-100 text-blue-700 dark:bg-blue-900/50 dark:text-blue-300': group.platform === 'openai',
                      'bg-purple-100 text-purple-700 dark:bg-purple-900/50 dark:text-purple-300': group.platform === 'gemini',
                      'bg-green-100 text-green-700 dark:bg-green-900/50 dark:text-green-300': group.platform === 'antigravity'
                    }">
                    {{ t(`admin.performance.concurrency.platforms.${group.platform}`) }}
                  </span>
                  <span class="text-xs text-gray-500 dark:text-gray-400 bg-white dark:bg-gray-800 px-2 py-0.5 rounded">
                    ({{ formatNumber(group.current_in_use) }} / {{ formatNumber(group.max_capacity) }})
                  </span>
                </div>
                <span class="text-sm font-bold"
                  :class="{
                    'text-green-600 dark:text-green-400': group.load_percentage < 50,
                    'text-yellow-600 dark:text-yellow-400': group.load_percentage >= 50 && group.load_percentage < 80,
                    'text-red-600 dark:text-red-400': group.load_percentage >= 80
                  }">
                  {{ group.load_percentage.toFixed(1) }}%
                </span>
              </div>
              <div class="h-3 overflow-hidden rounded-full bg-gray-200 dark:bg-gray-700">
                <div
                  class="h-full rounded-full transition-all duration-500"
                  :class="{
                    'bg-orange-500': group.platform === 'anthropic',
                    'bg-blue-500': group.platform === 'openai',
                    'bg-purple-500': group.platform === 'gemini',
                    'bg-green-500': group.platform === 'antigravity'
                  }"
                  :style="{ width: Math.min(group.load_percentage, 100) + '%' }"></div>
              </div>
              <div class="mt-2 flex items-center justify-between text-xs text-gray-500 dark:text-gray-400">
                <span>{{ t('admin.performance.concurrency.activeRequests') }}: {{ formatNumber(group.active_requests) }}</span>
                <span>{{ t('admin.performance.concurrency.waitingInQueue') }}: {{ formatNumber(group.waiting_in_queue) }}</span>
              </div>
            </div>
          </div>
          <div v-else class="text-center py-8">
            <svg class="h-12 w-12 text-gray-400 mx-auto mb-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
            </svg>
            <p class="text-sm text-gray-500 dark:text-gray-400">{{ t('common.noData') }}</p>
          </div>
        </div>

        <!-- Row 6: Overall Slot Usage -->
        <div class="card p-6">
          <h3 class="mb-4 text-base font-semibold text-gray-900 dark:text-white">
            {{ t('admin.performance.concurrency.title') }}
          </h3>
          <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
            <!-- User Slots -->
            <div class="bg-gradient-to-br from-blue-50 to-blue-100 dark:from-blue-900/20 dark:to-blue-900/10 rounded-lg p-4">
              <div class="flex items-center justify-between mb-3">
                <div class="flex items-center gap-2">
                  <div class="rounded-lg bg-blue-100 dark:bg-blue-900/30 p-1.5">
                    <svg class="h-4 w-4 text-blue-600 dark:text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                    </svg>
                  </div>
                  <span class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('admin.performance.concurrency.userSlots') }}</span>
                </div>
                <span class="text-lg font-bold text-blue-600 dark:text-blue-400">{{ metrics.slot_usage.user_slot }}</span>
              </div>
              <div class="h-3 overflow-hidden rounded-full bg-blue-200 dark:bg-blue-800/50">
                <div
                  class="h-full rounded-full bg-blue-500 transition-all duration-500"
                  :style="{ width: Math.min(metrics.slot_usage.user_slot / 10, 100) + '%' }"
                ></div>
              </div>
            </div>

            <!-- Account Slots -->
            <div class="bg-gradient-to-br from-green-50 to-green-100 dark:from-green-900/20 dark:to-green-900/10 rounded-lg p-4">
              <div class="flex items-center justify-between mb-3">
                <div class="flex items-center gap-2">
                  <div class="rounded-lg bg-green-100 dark:bg-green-900/30 p-1.5">
                    <svg class="h-4 w-4 text-green-600 dark:text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
                    </svg>
                  </div>
                  <span class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('admin.performance.concurrency.accountSlots') }}</span>
                </div>
                <span class="text-lg font-bold text-green-600 dark:text-green-400">{{ metrics.slot_usage.account_slot }}</span>
              </div>
              <div class="h-3 overflow-hidden rounded-full bg-green-200 dark:bg-green-800/50">
                <div
                  class="h-full rounded-full bg-green-500 transition-all duration-500"
                  :style="{ width: Math.min(metrics.slot_usage.account_slot / 50, 100) + '%' }"
                ></div>
              </div>
            </div>

            <!-- Wait Queue -->
            <div class="bg-gradient-to-br from-amber-50 to-amber-100 dark:from-amber-900/20 dark:to-amber-900/10 rounded-lg p-4">
              <div class="flex items-center justify-between mb-3">
                <div class="flex items-center gap-2">
                  <div class="rounded-lg bg-amber-100 dark:bg-amber-900/30 p-1.5">
                    <svg class="h-4 w-4 text-amber-600 dark:text-amber-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 0 0 1 8 0z" />
                    </svg>
                  </div>
                  <span class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('admin.performance.concurrency.waitQueue') }}</span>
                </div>
                <span class="text-lg font-bold text-amber-600 dark:text-amber-400">{{ metrics.slot_usage.wait_queue }}</span>
              </div>
              <div class="h-3 overflow-hidden rounded-full bg-amber-200 dark:bg-amber-800/50">
                <div
                  class="h-full rounded-full bg-amber-500 transition-all duration-500"
                  :style="{ width: Math.min(metrics.slot_usage.wait_queue / 5, 100) + '%' }"
                ></div>
              </div>
            </div>
          </div>
        </div>

        <!-- Uptime -->
        <div class="card p-4">
          <div class="flex items-center justify-between text-sm">
            <div class="flex items-center gap-3">
              <div class="flex items-center gap-2 text-gray-500 dark:text-gray-400">
                <svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 0 0 1 8 0z" />
                </svg>
                <span>{{ t('admin.performance.uptime') }}</span>
              </div>
            </div>
            <div class="flex items-center gap-4">
              <span class="text-gray-500 dark:text-gray-400">{{ t('admin.performance.lastUpdated') }}: {{ metrics.uptime.last_updated }}</span>
              <span class="font-mono text-gray-900 dark:text-white bg-gray-100 dark:bg-gray-800 px-3 py-1 rounded-lg">
                {{ metrics.uptime.start_time }}
              </span>
            </div>
          </div>
        </div>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { adminAPI } from '@/api/admin'
import type { PerformanceMetrics, HealthCheck } from '@/api/admin/performance'
import AppLayout from '@/components/layout/AppLayout.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(true)
const resetting = ref(false)
const enabling = ref(false)
const toggling = ref(false)
const enabled = ref(false)
const metrics = ref<PerformanceMetrics | null>(null)
const healthData = ref<HealthCheck | null>(null)
let refreshInterval: ReturnType<typeof setInterval> | null = null

const formatNumber = (value: number): string => {
  return value.toLocaleString()
}

const formatDuration = (ms: number): string => {
  if (ms >= 60000) {
    return `${(ms / 60000).toFixed(1)}m`
  } else if (ms >= 1000) {
    return `${(ms / 1000).toFixed(1)}s`
  }
  return `${Math.round(ms)}ms`
}

const loadEnabled = async () => {
  try {
    const res = await adminAPI.performance.getEnabled()
    enabled.value = res.enabled
  } catch (error) {
    console.error('Error loading enabled status:', error)
    enabled.value = false
  }
}

const loadData = async () => {
  if (!enabled.value) {
    loading.value = false
    return
  }
  
  loading.value = true
  try {
    const [metricsRes, healthRes] = await Promise.all([
      adminAPI.performance.getMetrics(),
      adminAPI.performance.healthCheck()
    ])
    metrics.value = metricsRes
    healthData.value = healthRes
  } catch (error) {
    appStore.showError(t('admin.performance.failedToLoad'))
    console.error('Error loading performance metrics:', error)
  } finally {
    loading.value = false
  }
}

const refreshData = async () => {
  await loadData()
}

const toggleEnabled = async () => {
  if (toggling.value) return

  const newValue = !enabled.value
  toggling.value = true

  try {
    await adminAPI.performance.setEnabled(newValue)
    enabled.value = newValue

    if (newValue) {
      appStore.showSuccess(t('admin.performance.enabledSuccess'))
      await loadData()
      startAutoRefresh()
    } else {
      appStore.showSuccess(t('admin.performance.disabledSuccess'))
      metrics.value = null
      healthData.value = null
      stopAutoRefresh()
    }
  } catch (error) {
    appStore.showError(t('admin.performance.toggleFailed'))
    console.error('Error toggling enabled status:', error)
  } finally {
    toggling.value = false
  }
}

const enableMonitoring = async () => {
  enabling.value = true
  try {
    await adminAPI.performance.setEnabled(true)
    enabled.value = true
    appStore.showSuccess(t('admin.performance.enabledSuccess'))
    await loadData()
    startAutoRefresh()
  } catch (error) {
    appStore.showError(t('admin.performance.enableFailed'))
    console.error('Error enabling monitoring:', error)
  } finally {
    enabling.value = false
  }
}

const resetMetrics = async () => {
  resetting.value = true
  try {
    await adminAPI.performance.resetMetrics()
    appStore.showSuccess(t('admin.performance.resetSuccess'))
    await loadData()
  } catch (error) {
    appStore.showError(t('admin.performance.resetFailed'))
    console.error('Error resetting metrics:', error)
  } finally {
    resetting.value = false
  }
}

const startAutoRefresh = () => {
  stopAutoRefresh()
  refreshInterval = setInterval(() => {
    loadData()
  }, 30000) // 每30秒自动刷新
}

const stopAutoRefresh = () => {
  if (refreshInterval) {
    clearInterval(refreshInterval)
    refreshInterval = null
  }
}

onMounted(async () => {
  await loadEnabled()
  await loadData()
  if (enabled.value) {
    startAutoRefresh()
  }
})

onUnmounted(() => {
  stopAutoRefresh()
})
</script>
