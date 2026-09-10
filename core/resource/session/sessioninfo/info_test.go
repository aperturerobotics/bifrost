package sessioninfo

import (
	"testing"

	provider "github.com/s4wave/spacewave/core/provider"
	api "github.com/s4wave/spacewave/core/provider/spacewave/api"
)

func TestShouldEmitOnboardingStatusFirstEmissionGate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		stateLoaded   bool
		accountStatus provider.ProviderAccountStatus
		expected      bool
	}{
		{
			name:          "holds none without account state",
			accountStatus: provider.ProviderAccountStatus_ProviderAccountStatus_NONE,
		},
		{
			name:          "holds pending without account state",
			accountStatus: provider.ProviderAccountStatus_ProviderAccountStatus_PENDING,
		},
		{
			name:          "holds ready without account state",
			accountStatus: provider.ProviderAccountStatus_ProviderAccountStatus_READY,
		},
		{
			name:          "emits unauthenticated without account state",
			accountStatus: provider.ProviderAccountStatus_ProviderAccountStatus_UNAUTHENTICATED,
			expected:      true,
		},
		{
			name:          "emits deleted without account state",
			accountStatus: provider.ProviderAccountStatus_ProviderAccountStatus_DELETED,
			expected:      true,
		},
		{
			name:          "emits dormant without account state",
			accountStatus: provider.ProviderAccountStatus_ProviderAccountStatus_DORMANT,
			expected:      true,
		},
		{
			name:          "emits failed without account state",
			accountStatus: provider.ProviderAccountStatus_ProviderAccountStatus_FAILED,
			expected:      true,
		},
		{
			name:          "emits once account state loaded",
			stateLoaded:   true,
			accountStatus: provider.ProviderAccountStatus_ProviderAccountStatus_PENDING,
			expected:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShouldEmitOnboardingStatus(tt.stateLoaded, tt.accountStatus)
			if got != tt.expected {
				t.Fatalf("expected %v, got %v", tt.expected, got)
			}
		})
	}
}

// TestBuildBillingUsageInfoPreservesSpendingConsent keeps exact ledger units and
// the payer's subscription boundary intact through the generated SDK projection.
func TestBuildBillingUsageInfoPreservesSpendingConsent(t *testing.T) {
	usage := BuildBillingUsageInfo(&api.BillingUsageResponse{
		StorageBytes: 123, StorageBaselineBytes: 107374182400,
		WriteOps: 50001, WriteOpsBaseline: 50000, ReadOps: 250000, ReadOpsBaseline: 250000,
		OverageLimitCents: 500, AccruedOverageMicrodollars: 20, ReservedOverageMicrodollars: 10,
		CurrentPeriodStart: 1776900000000, CurrentPeriodEnd: 1779492000000, OfferVersion: "cloud-monthly-v2",
		MonthlyPriceCents: 900, WriteMicrodollars: 30, ReadMicrodollars: 15, PolicyVersion: "2026-09-10",
	})
	if usage.GetMonthlyPriceCents() != 900 || usage.GetWriteMicrodollars() != 30 || usage.GetReadMicrodollars() != 15 || usage.GetPolicyVersion() != "2026-09-10" {
		t.Fatalf("billing projection lost accepted prices: %+v", usage)
	}
	if usage.GetAccruedOverageMicrodollars() != 20 || usage.GetOverageLimitCents() != 500 || usage.GetReservedOverageMicrodollars() != 10 {
		t.Fatalf("spending projection lost exact ledger units: %+v", usage)
	}
	if usage.GetCurrentPeriodStart() != 1776900000000 || usage.GetCurrentPeriodEnd() != 1779492000000 || usage.GetWriteOpsBaseline() != 50000 {
		t.Fatalf("spending projection lost its period or allowance: %+v", usage)
	}
}
