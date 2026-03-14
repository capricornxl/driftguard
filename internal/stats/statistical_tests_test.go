package stats

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKSTest(t *testing.T) {
	t.Run("identical distributions", func(t *testing.T) {
		sample1 := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
		sample2 := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

		result := KSTest(sample1, sample2)

		assert.Equal(t, 0.0, result.Statistic)
		assert.Equal(t, 1.0, result.PValue)
		assert.False(t, result.Significant)
	})

	t.Run("different distributions", func(t *testing.T) {
		sample1 := []float64{1, 2, 3, 4, 5}
		sample2 := []float64{6, 7, 8, 9, 10}

		result := KSTest(sample1, sample2)

		assert.Equal(t, 1.0, result.Statistic)
		assert.True(t, result.Significant)
	})

	t.Run("overlapping distributions", func(t *testing.T) {
		sample1 := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
		sample2 := []float64{5, 6, 7, 8, 9, 10, 11, 12, 13, 14}

		result := KSTest(sample1, sample2)

		assert.Greater(t, result.Statistic, 0.0)
		assert.Less(t, result.Statistic, 1.0)
	})

	t.Run("empty sample", func(t *testing.T) {
		sample1 := []float64{}
		sample2 := []float64{1, 2, 3}

		result := KSTest(sample1, sample2)

		assert.Equal(t, 0.0, result.Statistic)
		assert.Equal(t, 1.0, result.PValue)
		assert.False(t, result.Significant)
	})

	t.Run("real world latency drift", func(t *testing.T) {
		// Normal latency (ms)
		normal := []float64{95, 98, 102, 97, 100, 103, 99, 101, 96, 104}
		// Degraded latency (ms)
		degraded := []float64{150, 160, 155, 170, 165, 158, 162, 168, 153, 159}

		result := KSTest(normal, degraded)

		assert.Greater(t, result.Statistic, 0.5)
		assert.True(t, result.Significant)
	})
}

func TestCalculatePSI(t *testing.T) {
	t.Run("stable distribution", func(t *testing.T) {
		expected := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
		actual := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}

		result := CalculatePSI(expected, actual, 5)

		assert.Less(t, result.PSI, 0.1)
		assert.Equal(t, "stable", result.Stability)
	})

	t.Run("moderate shift", func(t *testing.T) {
		expected := []float64{1, 2, 3, 4, 5, 1, 2, 3, 4, 5, 1, 2, 3, 4, 5, 1, 2, 3, 4, 5}
		actual := []float64{3, 4, 5, 6, 7, 3, 4, 5, 6, 7, 3, 4, 5, 6, 7, 3, 4, 5, 6, 7}

		result := CalculatePSI(expected, actual, 5)

		assert.GreaterOrEqual(t, result.PSI, 0.1)
	})

	t.Run("significant shift", func(t *testing.T) {
		expected := []float64{1, 2, 3, 4, 5, 1, 2, 3, 4, 5, 1, 2, 3, 4, 5, 1, 2, 3, 4, 5}
		actual := []float64{15, 16, 17, 18, 19, 15, 16, 17, 18, 19, 15, 16, 17, 18, 19, 15, 16, 17, 18, 19}

		result := CalculatePSI(expected, actual, 5)

		assert.GreaterOrEqual(t, result.PSI, 0.2)
		assert.Equal(t, "significant", result.Stability)
	})

	t.Run("empty data", func(t *testing.T) {
		expected := []float64{}
		actual := []float64{1, 2, 3}

		result := CalculatePSI(expected, actual, 5)

		assert.Equal(t, 0.0, result.PSI)
		assert.Equal(t, "unknown", result.Stability)
	})
}

func TestAnalyzeTrend(t *testing.T) {
	t.Run("increasing trend", func(t *testing.T) {
		values := []float64{10, 20, 30, 40, 50, 60, 70}

		result := AnalyzeTrend(values)

		assert.Greater(t, result.Slope, 0)
		assert.Equal(t, "increasing", result.Direction)
		assert.Greater(t, result.Strength, 0.9)
		assert.True(t, result.Significant)
	})

	t.Run("decreasing trend", func(t *testing.T) {
		values := []float64{70, 60, 50, 40, 30, 20, 10}

		result := AnalyzeTrend(values)

		assert.Less(t, result.Slope, 0)
		assert.Equal(t, "decreasing", result.Direction)
		assert.Greater(t, result.Strength, 0.9)
		assert.True(t, result.Significant)
	})

	t.Run("stable trend", func(t *testing.T) {
		values := []float64{50, 51, 49, 50, 51, 49, 50}

		result := AnalyzeTrend(values)

		assert.Equal(t, "stable", result.Direction)
	})

	t.Run("insufficient data", func(t *testing.T) {
		values := []float64{50}

		result := AnalyzeTrend(values)

		assert.Equal(t, "unknown", result.Direction)
		assert.False(t, result.Significant)
	})
}

func TestDetectSpikes(t *testing.T) {
	t.Run("no spikes", func(t *testing.T) {
		values := []float64{100, 102, 98, 101, 99, 103, 97, 100}

		result := DetectSpikes(values, 2.0)

		assert.False(t, result.HasSpikes)
		assert.Empty(t, result.Spikes)
	})

	t.Run("with spikes", func(t *testing.T) {
		values := []float64{100, 102, 98, 101, 99, 500, 97, 100, -300}

		result := DetectSpikes(values, 2.0)

		assert.True(t, result.HasSpikes)
		assert.NotEmpty(t, result.Spikes)
		assert.Contains(t, result.Spikes, 5) // 500
		assert.Contains(t, result.Spikes, 8) // -300
	})

	t.Run("single spike", func(t *testing.T) {
		values := []float64{10, 11, 10, 11, 10, 100, 11, 10}

		result := DetectSpikes(values, 2.0)

		assert.True(t, result.HasSpikes)
		assert.Contains(t, result.Spikes, 5)
	})

	t.Run("insufficient data", func(t *testing.T) {
		values := []float64{100, 200}

		result := DetectSpikes(values, 2.0)

		assert.False(t, result.HasSpikes)
	})
}

func TestKSTestRealWorldScenarios(t *testing.T) {
	t.Run("agent response time regression", func(t *testing.T) {
		// Week 1: Normal operation
		week1 := generateNormalData(100, 500)
		// Week 2: Performance degradation
		week2 := generateNormalData(100, 750)

		result := KSTest(week1, week2)

		assert.True(t, result.Significant, "Should detect significant drift")
		assert.Greater(t, result.Statistic, 0.3)
	})

	t.Run("consistency score monitoring", func(t *testing.T) {
		// Baseline: High consistency
		baseline := generateNormalData(50, 0.95)
		// Current: Degrading consistency
		current := generateNormalData(50, 0.75)

		result := KSTest(baseline, current)

		assert.True(t, result.Significant)
	})
}

// Helper function to generate normally distributed data
func generateNormalData(n int, mean float64) []float64 {
	data := make([]float64, n)
	stdDev := mean * 0.1 // 10% standard deviation

	for i := 0; i < n; i++ {
		// Box-Muller transform for normal distribution
		u1 := float64(i+1) / float64(n)
		u2 := float64(i+2) / float64(n)
		z := math.Sqrt(-2*math.Log(u1)) * math.Cos(2*math.Pi*u2)
		data[i] = mean + z*stdDev
		if data[i] < 0 {
			data[i] = 0
		}
	}

	return data
}

func TestPSIBucketDetails(t *testing.T) {
	expected := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	actual := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	result := CalculatePSI(expected, actual, 5)

	assert.NotEmpty(t, result.Details)
	assert.Equal(t, 5, len(result.Details))

	// Verify bucket details are populated
	for _, detail := range result.Details {
		assert.NotEmpty(t, detail.Bucket)
		assert.GreaterOrEqual(t, detail.ExpectedPct, 0.0)
		assert.GreaterOrEqual(t, detail.ActualPct, 0.0)
	}
}
