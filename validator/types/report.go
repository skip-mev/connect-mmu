package types

const (
	GradeFailed = "FAIL"
	GradePassed = "PASS"

	StatusValid    = "VALID"
	StatusDegraded = "DEGRADED"
	StatusFailed   = "FAILED"
)

// Reports wraps an array of Report with summary info
type Reports struct {
	Valid    uint64   `json:"valid"`
	Degraded uint64   `json:"degraded"`
	Failed   uint64   `json:"failed"`
	Reports  []Report `json:"reports"`
}

type Report struct {
	Ticker          string           `json:"ticker"`
	Status          string           `json:"status"`
	PassingRatio    string           `json:"passing_ratio"`
	ReferencePrice  *float64         `json:"reference_price,omitempty"`
	ProviderReports []ProviderReport `json:"provider_reports"`
}

type ProviderReport struct {
	Name               string   `json:"name"`
	Grade              string   `json:"grade"`
	SuccessRate        float64  `json:"success_rate"`
	ZScore             float64  `json:"z_score"`
	AveragePrice       float64  `json:"average_price"`
	ReferencePriceDiff *float64 `json:"reference_price_jitter,omitempty"`
}
