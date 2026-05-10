package billing

import (
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

func TestComputeResponseImageToolTokenBasedBillingSnapshot(t *testing.T) {
	pricing := adminmodel.ResolvedModelPricing{
		Model:       "gpt-image-2",
		PriceUnit:   adminmodel.ProviderPriceUnitPer1KTokens,
		InputPrice:  0.008,
		OutputPrice: 0.03,
		Currency:    adminmodel.ProviderPriceCurrencyUSD,
	}

	snapshot, err := ComputeResponseImageToolTokenBasedBillingSnapshot(4160, pricing, 1)
	if err != nil {
		t.Fatalf("ComputeResponseImageToolTokenBasedBillingSnapshot() error = %v", err)
	}
	if snapshot.InputQuantity != 0 {
		t.Fatalf("InputQuantity = %v, want 0", snapshot.InputQuantity)
	}
	if snapshot.OutputQuantity != 4160 {
		t.Fatalf("OutputQuantity = %v, want 4160", snapshot.OutputQuantity)
	}
	if snapshot.InputAmount != 0 {
		t.Fatalf("InputAmount = %v, want 0", snapshot.InputAmount)
	}
	if snapshot.OutputAmount <= 0 {
		t.Fatalf("OutputAmount = %v, want > 0", snapshot.OutputAmount)
	}
}
