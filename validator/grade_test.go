package validator

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/skip-mev/connect-mmu/validator/types"
)

func ptrFloat64(f float64) *float64 {
	return &f
}

func TestCheckZScore(t *testing.T) {
	check := CheckZScore(1.5)

	tests := []struct {
		zScore float64
		want   bool
	}{
		{0.0, false},
		{1.0, false},
		{1.5, false},
		{1.6, true},
		{-1.5, false},
		{-1.6, true},
	}

	for _, tt := range tests {
		report := types.ProviderReport{ZScore: tt.zScore}
		got := check(report)
		require.Equal(t, tt.want, got, "CheckZScore(%v)", tt.zScore)
	}
}

func TestCheckSuccessThreshold(t *testing.T) {
	check := CheckSuccessThreshold(0.8) // bound at 0.8

	tests := []struct {
		successRate float64
		want        bool
	}{
		{1.0, false},
		{0.9, false},
		{0.8, false},
		{0.79, true},
		{0.0, true},
	}

	for _, tt := range tests {
		report := types.ProviderReport{SuccessRate: tt.successRate}
		got := check(report)
		require.Equal(t, tt.want, got, "CheckSuccessThreshold(%v)", tt.successRate)
	}
}

func TestCheckReferencePrice(t *testing.T) {
	check := CheckReferencePrice(0.05) // bound at 0.05

	value1 := 0.04
	value2 := 0.05
	value3 := 0.06

	tests := []struct {
		referencePriceDiff *float64
		want               bool
	}{
		{nil, false},
		{&value1, false},
		{&value2, false},
		{&value3, true},
	}

	for _, tt := range tests {
		report := types.ProviderReport{ReferencePriceDiff: tt.referencePriceDiff}
		got := check(report)
		require.Equal(t, tt.want, got, "CheckReferencePrice(%v)", tt.referencePriceDiff)
	}
}

func TestGradeReports(t *testing.T) {
	v := &Validator{}

	// Create provider reports
	pr1 := types.ProviderReport{
		Name:               "Provider1",
		ZScore:             1.0,
		SuccessRate:        0.9,
		AveragePrice:       100.0,
		ReferencePriceDiff: ptrFloat64(0.04),
	}
	pr2 := types.ProviderReport{
		Name:               "Provider2",
		ZScore:             2.0,
		SuccessRate:        0.7,
		AveragePrice:       101.0,
		ReferencePriceDiff: ptrFloat64(0.06),
	}
	pr3 := types.ProviderReport{
		Name:               "Provider3",
		ZScore:             -2.0,
		SuccessRate:        0.95,
		AveragePrice:       99.0,
		ReferencePriceDiff: nil,
	}

	report := types.Report{
		Ticker:          "TEST/USD",
		ProviderReports: []types.ProviderReport{pr1, pr2, pr3},
	}

	reports := []types.Report{report}

	zScoreCheck := CheckZScore(1.5)
	successRateCheck := CheckSuccessThreshold(0.8)
	referencePriceCheck := CheckReferencePrice(0.05)

	summary := v.GradeReports(reports, zScoreCheck, successRateCheck, referencePriceCheck)

	expectedGrades := []string{
		types.GradePassed, // pr1
		types.GradeFailed, // pr2 (fails on ZScore and SuccessRate)
		types.GradeFailed, // pr3 (fails on ZScore)
	}

	for i, providerReport := range summary.Reports[0].ProviderReports {
		require.Equal(t, expectedGrades[i], providerReport.Grade, "Provider %s grade", providerReport.Name)
	}

	expectedPassingRatio := "1/3"
	require.Equal(t, expectedPassingRatio, summary.Reports[0].PassingRatio, "PassingRatio")
}
