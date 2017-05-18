package logger

import (
	"bufio"
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
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		log.Info("%s\n", "abc")
	}
}
func Benchmark_WriteFile(b *testing.B) {
	file, _ := os.Create(".log")
	defer file.Close()
	defer os.Remove(".log")
	//writer := bufio.NewWriter(file)
	log := logger.NewLogger(logger.NewStreamHandler(file))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		log.Info("%s\n", "abc")
	}
}
func Benchmark_WriteBufioFile(b *testing.B) {
	file, _ := os.Create(".log")
	defer file.Close()
	defer os.Remove(".log")
	writer := bufio.NewWriter(file)
	log := logger.NewLogger(logger.NewStreamHandler(writer))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		log.Info("%s\n", "abc")
	}
}

func Benchmark_ParallWriteFile(b *testing.B) {
	file, _ := os.Create(".log")
	defer file.Close()
	defer os.Remove(".log")
	writer := bufio.NewWriter(file)
	log := logger.NewLogger(logger.NewStreamHandler(writer))
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			log.Info("%s\n", "abc")
		}
	})
}
