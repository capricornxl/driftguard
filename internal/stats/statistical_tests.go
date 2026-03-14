package stats

import (
	"math"
	"sort"
)

// KSTestResult holds the results of a Kolmogorov-Smirnov test
type KSTestResult struct {
	Statistic float64 // D statistic (maximum difference between CDFs)
	PValue    float64 // Approximate p-value
	Significant bool  // Whether the difference is statistically significant (p < 0.05)
}

// KS Test performs two-sample Kolmogorov-Smirnov test
// Returns D statistic and significance
// H0: Both samples come from the same distribution
// H1: Samples come from different distributions
func KSTest(sample1, sample2 []float64) KSTestResult {
	if len(sample1) == 0 || len(sample2) == 0 {
		return KSTestResult{
			Statistic:   0,
			PValue:      1,
			Significant: false,
		}
	}

	// Sort both samples
	sorted1 := make([]float64, len(sample1))
	sorted2 := make([]float64, len(sample2))
	copy(sorted1, sample1)
	copy(sorted2, sample2)
	sort.Float64s(sorted1)
	sort.Float64s(sorted2)

	// Calculate empirical CDFs and find maximum difference
	n1, n2 := float64(len(sample1)), float64(len(sample2))
	maxDiff := 0.0
	i, j := 0, 0

	for i < len(sorted1) && j < len(sorted2) {
		var x float64
		if sorted1[i] < sorted2[j] {
			x = sorted1[i]
			i++
		} else {
			x = sorted2[j]
			j++
		}

		// Calculate proportion of values <= x in each sample
		f1 := float64(i) / n1
		f2 := float64(j) / n2

		diff := math.Abs(f1 - f2)
		if diff > maxDiff {
			maxDiff = diff
		}
	}

	// Calculate approximate p-value using Kolmogorov distribution
	// For large samples, use asymptotic approximation
	n := (n1 * n2) / (n1 + n2)
	lambda := maxDiff * math.Sqrt(n)

	// Approximate p-value using Kolmogorov distribution formula
	// P(D > d) ≈ 2 * exp(-2 * n * d^2) for large n
	pValue := 2 * math.Exp(-2*n*maxDiff*maxDiff)
	if pValue > 1 {
		pValue = 1
	}

	return KSTestResult{
		Statistic:   maxDiff,
		PValue:      pValue,
		Significant: pValue < 0.05,
	}
}

// PSICalculation holds PSI calculation results
type PSICalculation struct {
	PSI       float64 // Population Stability Index
	Stability string  // "stable", "moderate", "significant"
	Details   []BucketDetail
}

// BucketDetail contains details for each bucket in PSI calculation
type BucketDetail struct {
	Bucket         string
	ExpectedCount  int
	ActualCount    int
	ExpectedPct    float64
	ActualPct      float64
	PSIContribution float64
}

// CalculatePSI computes Population Stability Index
// PSI < 0.1: Stable
// 0.1 <= PSI < 0.2: Moderate shift
// PSI >= 0.2: Significant shift
func CalculatePSI(expected, actual []float64, numBuckets int) PSICalculation {
	if len(expected) == 0 || len(actual) == 0 {
		return PSICalculation{
			PSI:       0,
			Stability: "unknown",
			Details:   []BucketDetail{},
		}
	}

	// Create buckets based on expected distribution
	buckets := createBuckets(expected, numBuckets)

	// Count values in each bucket
	expectedCounts := make([]int, numBuckets)
	actualCounts := make([]int, numBuckets)

	for _, v := range expected {
		idx := getBucketIndex(v, buckets)
		if idx >= 0 && idx < numBuckets {
			expectedCounts[idx]++
		}
	}

	for _, v := range actual {
		idx := getBucketIndex(v, buckets)
		if idx >= 0 && idx < numBuckets {
			actualCounts[idx]++
		}
	}

	// Calculate PSI
	totalExpected := float64(len(expected))
	totalActual := float64(len(actual))
	
	var psi float64
	details := make([]BucketDetail, numBuckets)

	for i := 0; i < numBuckets; i++ {
		expectedPct := float64(expectedCounts[i]) / totalExpected
		actualPct := float64(actualCounts[i]) / totalActual

		// Add small epsilon to avoid log(0)
		if expectedPct < 0.0001 {
			expectedPct = 0.0001
		}
		if actualPct < 0.0001 {
			actualPct = 0.0001
		}

		psiContribution := (actualPct - expectedPct) * math.Log(actualPct/expectedPct)
		psi += psiContribution

		details[i] = BucketDetail{
			Bucket:         buckets[i],
			ExpectedCount:  expectedCounts[i],
			ActualCount:    actualCounts[i],
			ExpectedPct:    expectedPct * 100,
			ActualPct:      actualPct * 100,
			PSIContribution: psiContribution,
		}
	}

	// Determine stability level
	stability := "stable"
	if psi >= 0.2 {
		stability = "significant"
	} else if psi >= 0.1 {
		stability = "moderate"
	}

	return PSICalculation{
		PSI:       psi,
		Stability: stability,
		Details:   details,
	}
}

// Helper functions

func createBuckets(data []float64, numBuckets int) []string {
	if len(data) == 0 {
		return []string{}
	}

	// Find min and max
	minVal, maxVal := data[0], data[0]
	for _, v := range data {
		if v < minVal {
			minVal = v
		}
		if v > maxVal {
			maxVal = v
		}
	}

	// Create bucket boundaries
	buckets := make([]string, numBuckets)
	step := (maxVal - minVal) / float64(numBuckets)

	for i := 0; i < numBuckets; i++ {
		lower := minVal + float64(i)*step
		upper := minVal + float64(i+1)*step
		if i == numBuckets-1 {
			buckets[i] = formatRange(lower, upper, true)
		} else {
			buckets[i] = formatRange(lower, upper, false)
		}
	}

	return buckets
}

func getBucketIndex(value float64, buckets []string) int {
	// Parse bucket boundaries and find the right bucket
	// This is a simplified implementation
	for i, bucket := range buckets {
		if valueInBucket(value, bucket) {
			return i
		}
	}
	return -1
}

func valueInBucket(value float64, bucket string) bool {
	// Simplified bucket checking - in production, parse the bucket string
	// For now, assume equal-width buckets
	return true
}

func formatRange(lower, upper float64, includeUpper bool) string {
	if includeUpper {
		return formatNumber(lower) + " - " + formatNumber(upper)
	}
	return formatNumber(lower) + " - " + formatNumber(upper) + " (excl)"
}

func formatNumber(n float64) string {
	if n == float64(int(n)) {
		return string(rune(int(n)))
	}
	return string(rune(int(n*100)/100))
}

// TrendAnalysis analyzes the trend of a time series
type TrendAnalysis struct {
	Slope       float64 // Rate of change per time unit
	Direction   string  // "increasing", "decreasing", "stable"
	Strength    float64 // R-squared value (0-1)
	Significant bool    // Whether the trend is statistically significant
}

// AnalyzeTrend performs linear regression to detect trends
func AnalyzeTrend(values []float64) TrendAnalysis {
	n := len(values)
	if n < 2 {
		return TrendAnalysis{
			Slope:       0,
			Direction:   "unknown",
			Strength:    0,
			Significant: false,
		}
	}

	// Calculate means
	var sumX, sumY, sumXY, sumX2, sumY2 float64
	for i, y := range values {
		x := float64(i)
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
		sumY2 += y * y
	}

	meanX := sumX / float64(n)
	meanY := sumY / float64(n)

	// Calculate slope and intercept
	denominator := sumX2 - float64(n)*meanX*meanX
	if denominator == 0 {
		return TrendAnalysis{
			Slope:       0,
			Direction:   "stable",
			Strength:    0,
			Significant: false,
		}
	}

	slope := (sumXY - float64(n)*meanX*meanY) / denominator
	intercept := meanY - slope*meanX

	// Calculate R-squared
	var ssTot, ssRes float64
	for i, y := range values {
		x := float64(i)
		predicted := slope*x + intercept
		ssRes += (y - predicted) * (y - predicted)
		ssTot += (y - meanY) * (y - meanY)
	}

	rSquared := 1 - (ssRes / ssTot)
	if rSquared < 0 {
		rSquared = 0
	}

	// Determine direction
	direction := "stable"
	if slope > 0.01 {
		direction = "increasing"
	} else if slope < -0.01 {
		direction = "decreasing"
	}

	// Determine significance (simple heuristic)
	significant := rSquared > 0.5 && n >= 5

	return TrendAnalysis{
		Slope:       slope,
		Direction:   direction,
		Strength:    rSquared,
		Significant: significant,
	}
}

// SpikeDetection detects sudden spikes in data
type SpikeResult struct {
	HasSpikes   bool
	Spikes      []int    // Indices where spikes occur
	Threshold   float64  // Threshold used for detection
	Mean        float64
	StdDev      float64
}

// DetectSpikes identifies values that are more than threshold standard deviations from mean
func DetectSpikes(values []float64, threshold float64) SpikeResult {
	if len(values) < 3 {
		return SpikeResult{
			HasSpikes: false,
			Spikes:    []int{},
			Threshold: threshold,
		}
	}

	// Calculate mean
	var sum float64
	for _, v := range values {
		sum += v
	}
	mean := sum / float64(len(values))

	// Calculate standard deviation
	var sumSq float64
	for _, v := range values {
		diff := v - mean
		sumSq += diff * diff
	}
	stdDev := math.Sqrt(sumSq / float64(len(values)-1))

	// Detect spikes
	var spikes []int
	lowerBound := mean - threshold*stdDev
	upperBound := mean + threshold*stdDev

	for i, v := range values {
		if v < lowerBound || v > upperBound {
			spikes = append(spikes, i)
		}
	}

	return SpikeResult{
		HasSpikes: len(spikes) > 0,
		Spikes:    spikes,
		Threshold: threshold,
		Mean:      mean,
		StdDev:    stddev,
	}
}
