package service

import (
	"context"
	"log"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/google/wire"
)

// BuildInfo contains build information
type BuildInfo struct {
	Version   string
	BuildType string
}

// ProvidePricingService creates and initializes PricingService
func ProvidePricingService(cfg *config.Config, remoteClient PricingRemoteClient) (*PricingService, error) {
	svc := NewPricingService(cfg, remoteClient)
	if err := svc.Initialize(); err != nil {
		// Pricing service initialization failure should not block startup, use fallback prices
		println("[Service] Warning: Pricing service initialization failed:", err.Error())
	}
	return svc, nil
}

// ProvideUpdateService creates UpdateService with BuildInfo
func ProvideUpdateService(cache UpdateCache, githubClient GitHubReleaseClient, buildInfo BuildInfo) *UpdateService {
	return NewUpdateService(cache, githubClient, buildInfo.Version, buildInfo.BuildType)
}

// ProvideEmailQueueService creates EmailQueueService with default worker count
func ProvideEmailQueueService(emailService *EmailService) *EmailQueueService {
	return NewEmailQueueService(emailService, 3)
}

// ProvideTokenRefreshService creates and starts TokenRefreshService
func ProvideTokenRefreshService(
	accountRepo AccountRepository,
	oauthService *OAuthService,
	openaiOAuthService *OpenAIOAuthService,
	geminiOAuthService *GeminiOAuthService,
	antigravityOAuthService *AntigravityOAuthService,
	cfg *config.Config,
) *TokenRefreshService {
	svc := NewTokenRefreshService(accountRepo, oauthService, openaiOAuthService, geminiOAuthService, antigravityOAuthService, cfg)
	svc.Start()
	return svc
}

// ProvideTimingWheelService creates and starts TimingWheelService
func ProvideTimingWheelService() *TimingWheelService {
	svc := NewTimingWheelService()
	svc.Start()
	return svc
}

// ProvideDeferredService creates and starts DeferredService
func ProvideDeferredService(accountRepo AccountRepository, timingWheel *TimingWheelService) *DeferredService {
	svc := NewDeferredService(accountRepo, timingWheel, 10*time.Second)
	svc.Start()
	return svc
}

// ProvideConcurrencyService creates ConcurrencyService and starts slot cleanup worker.
func ProvideConcurrencyService(cache ConcurrencyCache, accountRepo AccountRepository, cfg *config.Config) *ConcurrencyService {
	svc := NewConcurrencyService(cache)
	if cfg != nil {
		svc.StartSlotCleanupWorker(accountRepo, cfg.Gateway.Scheduling.SlotCleanupInterval)
	}
	return svc
}

// ProvidePerformanceMonitor creates PerformanceMonitor and starts slot usage update worker.
func ProvidePerformanceMonitor(concurrencyService *ConcurrencyService, accountRepo AccountRepository, groupRepo GroupRepository) *PerformanceMonitor {
	monitor := NewPerformanceMonitor()

	// Start background worker to update slot usage every 5 seconds
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

			// Get all schedulable accounts
			accounts, err := accountRepo.ListSchedulable(ctx)
			if err != nil {
				log.Printf("[PerformanceMonitor] Warning: list schedulable accounts failed: %v", err)
				cancel()
				continue
			}

			// Calculate total concurrent slots in use and group by platform
			var totalUserSlots int64
			var totalAccountSlots int64
			var totalWaitQueue int64

			// Group by platform: anthropic, openai, gemini, antigravity
			// Initialize all platforms with zero values
			concurrencyByPlatform := map[string]map[string]int64{
				PlatformAnthropic:   {"current": 0, "max": 0, "waiting": 0, "active": 0},
				PlatformOpenAI:      {"current": 0, "max": 0, "waiting": 0, "active": 0},
				PlatformGemini:      {"current": 0, "max": 0, "waiting": 0, "active": 0},
				PlatformAntigravity: {"current": 0, "max": 0, "waiting": 0, "active": 0},
			}

			// Group by group
			concurrencyByGroup := make(map[int64]map[string]any)
			// Initialize empty groups map with group info
			allGroups, err := groupRepo.ListActive(ctx)
			if err != nil {
				log.Printf("[PerformanceMonitor] Warning: list active groups failed: %v", err)
			} else {
				for _, group := range allGroups {
					concurrencyByGroup[group.ID] = map[string]any{
						"group_name": group.Name,
						"platform":   group.Platform,
						"current":    int64(0),
						"max":        int64(0),
						"waiting":    int64(0),
					}
				}
			}

			for _, account := range accounts {
				// Skip if platform is not recognized
				platform := account.Platform
				if platform != PlatformAnthropic && platform != PlatformOpenAI &&
					platform != PlatformGemini && platform != PlatformAntigravity {
					log.Printf("[PerformanceMonitor] Warning: unknown platform '%s' for account %d", platform, account.ID)
					continue
				}

				// Get current concurrency for this account
				accountCount, err := concurrencyService.cache.GetAccountConcurrency(ctx, account.ID)
				if err == nil {
					totalAccountSlots += int64(accountCount)
					concurrencyByPlatform[platform]["current"] += int64(accountCount)
				}

				// Get account wait count
				waitCount, err := concurrencyService.cache.GetAccountWaitingCount(ctx, account.ID)
				if err == nil {
					totalWaitQueue += int64(waitCount)
					concurrencyByPlatform[platform]["waiting"] += int64(waitCount)
				}

				// Add max capacity for this platform
				concurrencyByPlatform[platform]["max"] += int64(account.Concurrency)

				// Aggregate by group
				for _, group := range account.Groups {
					if groupData, exists := concurrencyByGroup[group.ID]; exists {
						if current, ok := groupData["current"].(int64); ok {
							groupData["current"] = current + int64(accountCount)
						}
						if maxCap, ok := groupData["max"].(int64); ok {
							groupData["max"] = maxCap + int64(account.Concurrency)
						}
						if waiting, ok := groupData["waiting"].(int64); ok {
							groupData["waiting"] = waiting + int64(waitCount)
						}
					}
				}
			}

			// For user slots, we approximate based on account concurrency
			totalUserSlots = totalAccountSlots

			// Update simple slot usage
			monitor.UpdateSlotUsage(totalUserSlots, totalAccountSlots, totalWaitQueue)

			// Update concurrency by platform - pass the entire map at once
			monitor.UpdateConcurrencyByPlatform(concurrencyByPlatform)

			// Update concurrency by group
			monitor.UpdateConcurrencyByGroup(concurrencyByGroup)

			cancel()
		}
	}()

	return monitor
}

// ProviderSet is the Wire provider set for all services
var ProviderSet = wire.NewSet(
	// Core services
	NewAuthService,
	NewUserService,
	NewAPIKeyService,
	NewGroupService,
	NewAccountService,
	NewProxyService,
	NewRedeemService,
	NewUsageService,
	NewDashboardService,
	ProvidePricingService,
	NewBillingService,
	NewBillingCacheService,
	NewAdminService,
	NewGatewayService,
	NewOpenAIGatewayService,
	NewOAuthService,
	NewOpenAIOAuthService,
	NewGeminiOAuthService,
	NewGeminiQuotaService,
	NewAntigravityOAuthService,
	NewGeminiTokenProvider,
	NewGeminiMessagesCompatService,
	NewAntigravityTokenProvider,
	NewAntigravityGatewayService,
	NewRateLimitService,
	NewAccountUsageService,
	NewAccountTestService,
	NewSettingService,
	NewEmailService,
	ProvideEmailQueueService,
	NewTurnstileService,
	NewSubscriptionService,
	ProvideConcurrencyService,
	NewIdentityService,
	NewCRSSyncService,
	ProvideUpdateService,
	ProvideTokenRefreshService,
	ProvideTimingWheelService,
	ProvideDeferredService,
	NewAntigravityQuotaFetcher,
	NewUserAttributeService,
	NewUsageCache,
	ProvidePerformanceMonitor,
)
