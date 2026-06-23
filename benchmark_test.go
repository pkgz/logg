package logg

import (
	"io"
	"log"
	"testing"
)

var benchmarkMessages = map[string][]byte{
	"long message":  []byte("Lorem Ipsum is simply dummy text of the printing and typesetting industry. Lorem Ipsum has been the industry's standard dummy text ever since the 1500s, when an unknown printer took a galley of type and scrambled it to make a type specimen book. It has survived not only five centuries, but also the leap into electronic typesetting, remaining essentially unchanged. It was popularised in the 1960s with the release of Letraset sheets containing Lorem Ipsum passages, and more recently with desktop publishing software like Aldus PageMaker including versions of Lorem Ipsum."),
	"short message": []byte("test logging, but use a somewhat realistic message length."),
	"short level":   []byte("INF test logging, but use a somewhat realistic message length."),
	"long level":    []byte("ERROR test logging, but use a somewhat realistic message length."),
}

func BenchmarkLog_Print(b *testing.B) {
	log.SetOutput(io.Discard)

	for name, tc := range benchmarkMessages {
		str := string(tc)

		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()

			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					log.Print(str)
				}
			})
		})
	}
}

func BenchmarkLogg_Log_Print(b *testing.B) {
	NewGlobal(io.Discard)

	for name, tc := range benchmarkMessages {
		str := string(tc)

		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()

			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					log.Print(str)
				}
			})
		})
	}
}

func BenchmarkLogg_Write_Pretty(b *testing.B) {
	logger := New(io.Discard)

	for name, tc := range benchmarkMessages {
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()

			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					_, _ = logger.Write(tc)
				}
			})
		})
	}
}

func BenchmarkLogg_Write_Json(b *testing.B) {
	logger := New(io.Discard)
	logger.format = Json

	for name, tc := range benchmarkMessages {
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()

			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					_, _ = logger.Write(tc)
				}
			})
		})
	}
}

func BenchmarkLogg_Print(b *testing.B) {
	logger := New(io.Discard)

	for name, tc := range benchmarkMessages {
		str := string(tc)

		b.Run(name, func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()

			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					logger.Print(str)
				}
			})
		})
	}
}

func BenchmarkLogg_Printf(b *testing.B) {
	logger := New(io.Discard)

	for name, tc := range benchmarkMessages {
		str := string(tc)

		b.Run(name, func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()

			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					logger.Printf("%s", str)
				}
			})
		})
	}
}
