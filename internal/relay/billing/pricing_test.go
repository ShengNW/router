package billing

import (
	"math"
	"testing"

	adminmodel "github.com/yeying-community/router/internal/admin/model"
)

func TestResolveImageBillingMode(t *testing.T) {
	tests := []struct {
		name    string
		pricing adminmodel.ResolvedModelPricing
		want    ImageBillingMode
	}{
		{
			name: "per image",
			pricing: adminmodel.ResolvedModelPricing{
				PriceUnit: adminmodel.ProviderPriceUnitPerImage,
			},
			want: ImageBillingModePerImage,
		},
		{
			name: "per request",
			pricing: adminmodel.ResolvedModelPricing{
				PriceUnit: adminmodel.ProviderPriceUnitPerRequest,
			},
			want: ImageBillingModePerCall,
		},
		{
			name: "per task",
			pricing: adminmodel.ResolvedModelPricing{
				PriceUnit: adminmodel.ProviderPriceUnitPerTask,
			},
			want: ImageBillingModePerCall,
		},
		{
			name: "token based",
			pricing: adminmodel.ResolvedModelPricing{
				PriceUnit: adminmodel.ProviderPriceUnitPer1KTokens,
			},
			want: ImageBillingModeTokenBased,
		},
		{
			name: "unknown",
			pricing: adminmodel.ResolvedModelPricing{
				PriceUnit: "per_pixel",
			},
			want: ImageBillingModeUnsupported,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ResolveImageBillingMode(tt.pricing); got != tt.want {
				t.Fatalf("ResolveImageBillingMode() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestComputeImageBillingSnapshotByMode(t *testing.T) {
	t.Run("per image uses image count and multiplier", func(t *testing.T) {
		pricing := adminmodel.ResolvedModelPricing{
			Model:      "dall-e-3",
			PriceUnit:  adminmodel.ProviderPriceUnitPerImage,
			InputPrice: 0.04,
			Currency:   adminmodel.ProviderPriceCurrencyUSD,
		}

		snapshot, err := ComputeImageBillingSnapshot(2, 1.5, pricing, 1)
		if err != nil {
			t.Fatalf("ComputeImageBillingSnapshot() error = %v", err)
		}
		if snapshot.InputQuantity != 3 {
			t.Fatalf("InputQuantity = %v, want 3", snapshot.InputQuantity)
		}
		if snapshot.InputAmount != 0.12 {
			t.Fatalf("InputAmount = %v, want 0.12", snapshot.InputAmount)
		}
	})

	t.Run("per call ignores image count multiplier", func(t *testing.T) {
		pricing := adminmodel.ResolvedModelPricing{
			Model:      "foo-image",
			PriceUnit:  adminmodel.ProviderPriceUnitPerRequest,
			InputPrice: 0.5,
			Currency:   adminmodel.ProviderPriceCurrencyUSD,
		}

		snapshot, err := ComputeImageBillingSnapshot(4, 3, pricing, 1)
		if err != nil {
			t.Fatalf("ComputeImageBillingSnapshot() error = %v", err)
		}
		if snapshot.InputQuantity != 1 {
			t.Fatalf("InputQuantity = %v, want 1", snapshot.InputQuantity)
		}
		if snapshot.InputAmount != 0.5 {
			t.Fatalf("InputAmount = %v, want 0.5", snapshot.InputAmount)
		}
	})

	t.Run("token based returns explicit error", func(t *testing.T) {
		pricing := adminmodel.ResolvedModelPricing{
			Model:      "gpt-image-2",
			PriceUnit:  adminmodel.ProviderPriceUnitPer1KTokens,
			InputPrice: 0.008,
			Currency:   adminmodel.ProviderPriceCurrencyUSD,
		}

		if _, err := ComputeImageBillingSnapshot(1, 1, pricing, 1); err == nil {
			t.Fatal("ComputeImageBillingSnapshot() error = nil, want error")
		}
	})
}

func TestComputeTraditionalImageTokenBasedBillingSnapshot(t *testing.T) {
	pricing := adminmodel.ResolvedModelPricing{
		Model:       "gpt-image-2",
		PriceUnit:   adminmodel.ProviderPriceUnitPer1KTokens,
		InputPrice:  0.008,
		OutputPrice: 0.03,
		Currency:    adminmodel.ProviderPriceCurrencyUSD,
	}

	snapshot, err := ComputeTraditionalImageTokenBasedBillingSnapshot(100, 1056, pricing, 1)
	if err != nil {
		t.Fatalf("ComputeTraditionalImageTokenBasedBillingSnapshot() error = %v", err)
	}
	if snapshot.InputQuantity != 100 {
		t.Fatalf("InputQuantity = %v, want 100", snapshot.InputQuantity)
	}
	if snapshot.OutputQuantity != 1056 {
		t.Fatalf("OutputQuantity = %v, want 1056", snapshot.OutputQuantity)
	}
	if snapshot.Amount <= 0 {
		t.Fatalf("Amount = %v, want > 0", snapshot.Amount)
	}
	if snapshot.YYCAmount <= 0 {
		t.Fatalf("YYCAmount = %d, want > 0", snapshot.YYCAmount)
	}
}

func TestComputeTokenBasedBillingSnapshot(t *testing.T) {
	pricing := adminmodel.ResolvedModelPricing{
		Model:       "gpt-image-2",
		PriceUnit:   adminmodel.ProviderPriceUnitPer1KTokens,
		InputPrice:  0.005,
		OutputPrice: 0.03,
		Currency:    adminmodel.ProviderPriceCurrencyUSD,
	}

	snapshot, err := ComputeTokenBasedBillingSnapshot(100, 7033.333333333333, pricing, 1)
	if err != nil {
		t.Fatalf("ComputeTokenBasedBillingSnapshot() error = %v", err)
	}
	if snapshot.InputAmount != 0.0005 {
		t.Fatalf("InputAmount = %v, want 0.0005", snapshot.InputAmount)
	}
	if math.Abs(snapshot.OutputAmount-0.211) > 1e-9 {
		t.Fatalf("OutputAmount = %v, want 0.211", snapshot.OutputAmount)
	}
}

func TestComputeResponseImageToolTokenBasedBillingSnapshot(t *testing.T) {
	pricing := adminmodel.ResolvedModelPricing{
		Model:       "gpt-image-2",
		PriceUnit:   adminmodel.ProviderPriceUnitPer1KTokens,
		InputPrice:  0.008,
		OutputPrice: 0.03,
		Currency:    adminmodel.ProviderPriceCurrencyUSD,
	}

	snapshot, err := ComputeResponseImageToolTokenBasedBillingSnapshot(7033.333333333333, pricing, 1)
	if err != nil {
		t.Fatalf("ComputeResponseImageToolTokenBasedBillingSnapshot() error = %v", err)
	}
	if snapshot.InputQuantity != 0 {
		t.Fatalf("InputQuantity = %v, want 0", snapshot.InputQuantity)
	}
	if math.Abs(snapshot.OutputQuantity-7033.333333333333) > 1e-9 {
		t.Fatalf("OutputQuantity = %v, want %v", snapshot.OutputQuantity, 7033.333333333333)
	}
	if snapshot.InputAmount != 0 {
		t.Fatalf("InputAmount = %v, want 0", snapshot.InputAmount)
	}
	if math.Abs(snapshot.OutputAmount-0.211) > 1e-9 {
		t.Fatalf("OutputAmount = %v, want 0.211", snapshot.OutputAmount)
	}
}

func TestComputeExplicitAmountBillingSnapshot(t *testing.T) {
	pricing := adminmodel.ResolvedModelPricing{
		Model:     "gpt-image-2",
		PriceUnit: adminmodel.ProviderPriceUnitPer1KTokens,
		Currency:  adminmodel.ProviderPriceCurrencyUSD,
	}
	snapshot, err := ComputeExplicitAmountBillingSnapshot(4454, 7033.333333333333, 0.035332, 0.211, pricing, 1, true)
	if err != nil {
		t.Fatalf("ComputeExplicitAmountBillingSnapshot() error = %v", err)
	}
	if math.Abs(snapshot.InputAmount-0.035332) > 1e-9 {
		t.Fatalf("InputAmount = %v, want 0.035332", snapshot.InputAmount)
	}
	if math.Abs(snapshot.OutputAmount-0.211) > 1e-9 {
		t.Fatalf("OutputAmount = %v, want 0.211", snapshot.OutputAmount)
	}
	if math.Abs(snapshot.Amount-0.246332) > 1e-9 {
		t.Fatalf("Amount = %v, want 0.246332", snapshot.Amount)
	}
}
