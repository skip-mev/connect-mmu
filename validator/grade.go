package validator

import (
	"fmt"
	"slices"

	"github.com/skip-mev/connect-mmu/validator/types"
)

// CheckFailed checks a provider report and returns "true" if the report should be considered failed.
// return false if the report passed.
type CheckFailed func(report types.ProviderReport) bool

func CheckZScore(bound float64) CheckFailed {
	return func(report types.ProviderReport) bool {
		return report.ZScore > bound || report.ZScore < -bound
	}
}

func CheckSuccessThreshold(bound float64) CheckFailed {
	return func(report types.ProviderReport) bool {
		return report.SuccessRate < bound
	}
}

func CheckReferencePrice(bound float64) CheckFailed {
	return func(report types.ProviderReport) bool {
		if report.ReferencePriceDiff == nil {
			return false
		}
		return *report.ReferencePriceDiff > bound
	}
}

// GradeReports will run checks on the reports, and mark them as failed/passed.
// It will also update the main report indicating the ratio of successful providers (e.g. 3/5).
func (v *Validator) GradeReports(reports []types.Report, failChecks ...CheckFailed) types.Reports {
	var (
		numValid    uint64
		numDegraded uint64
		numFailed   uint64
	)

	for i, report := range reports {
		passed := 0
		for j, providerReport := range report.ProviderReports {
			for _, check := range failChecks {
				if check(providerReport) {
					providerReport.Grade = types.GradeFailed
					break
				}
			}
			if providerReport.Grade == "" {
				providerReport.Grade = types.GradePassed
				passed++
			}
			report.ProviderReports[j] = providerReport
		}
		report.PassingRatio = fmt.Sprintf("%d/%d", passed, len(report.ProviderReports))

		switch {
		case passed == len(report.ProviderReports):
			report.Status = types.StatusValid
			numValid++
		case len(report.ProviderReports) > passed && passed > 0:
			report.Status = types.StatusDegraded
			numDegraded++
		default:
			report.Status = types.StatusFailed
			numFailed++
		}

		reports[i] = report
	}

	// sort so that all FAILED markets are first
	slices.SortFunc(reports, func(a, _ types.Report) int {
		switch a.Status {
		case types.StatusFailed:
			return -1
		default:
			return 0
		}
	})

	return types.Reports{
		Reports:  reports,
		Valid:    numValid,
		Degraded: numDegraded,
		Failed:   numFailed,
	}
}
