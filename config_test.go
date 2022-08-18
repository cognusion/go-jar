package jar

import (
	"testing"
)

func Benchmark_ConfigGetEmpty(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Conf.GetString("thiswillneverexist")
	}
}

func Benchmark_ConfigGetString(b *testing.B) {
	Conf.Set("thisexists", "This Value Is Cool")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Conf.GetString("thisexists")
	}
}
