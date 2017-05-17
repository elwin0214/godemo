package logger

import (
	"logger"
	"os"
	"testing"
)

//go test -test.bench=".*"  src/logger/logger_test.go
//go test src/logger/logger_test.go
func Test_Log(t *testing.T) {

	handler := logger.NewStreamHandler(os.Stdout)
	log := logger.NewLogger(handler)
	log.SetLevel(logger.LevelTrace)
	log.Info("%s\n", "abc")
	log.Error("%s\n", "abc")

}

func Benchmark_Log(b *testing.B) {
	log := logger.NewLogger(logger.NewNullHandler())
	for i := 0; i < b.N; i++ {
		log.Info("%s\n", "abc")
	}
}
