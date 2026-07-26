package initializr_test

import (
	"testing"

	"github.com/saireddy-shyamakura/springx/internal/initializr"
)

// BenchmarkBuildURL measures URL construction overhead.
func BenchmarkBuildURL(b *testing.B) {
	cfg := minimalCfg()
	cfg.Dependencies = []string{"web", "data-jpa", "postgresql", "lombok", "actuator", "validation"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = initializr.BuildURL(cfg)
	}
}
