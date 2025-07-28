package Final_Task

import (
	"awesomeProject1/Final_Task"
	"bytes"
	"context"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func testLogger() *log.Logger {
	return log.New(io.Discard, "[TEST] ", log.LstdFlags) // Лог не выводится
}

func createTempFile(t testing.TB, content string) string {
	tmpFile := filepath.Join(t.TempDir(), "temp.txt")
	err := os.WriteFile(tmpFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Ошибка создания файла: %v", err)
	}
	return tmpFile
}
func TestWriteDatas(t *testing.T) {
	cases := []struct {
		name string
		src  string
		data []byte
	}{
		{
			name: "nums",
			src:  "test1",
			data: []byte{1, 2, 3},
		}, {
			name: "txt",
			src:  "test2",
			data: []byte("hello"),
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			logger := testLogger()

			src := createTempFile(t, "")
			data := []byte("test write\n")

			var wg sync.WaitGroup
			wg.Add(1)
			go main.writeDatas(ctx, &wg, src, data, logger)
			wg.Wait()

			readData, _ := os.ReadFile(src)
			if !bytes.Contains(readData, data) {
				t.Errorf("Ожидалось, что файл содержит '%s', но он содержит: %s", data, readData)
			}
		})
	}
}

func TestCopyFile(t *testing.T) {
	cases := []struct {
		name string
		src  string
		dst  string
		data []byte
	}{
		{name: "test1", src: "dst1", dst: "src1", data: []byte{1, 2, 3}},
		{name: "test2", src: "dst2", dst: "src2", data: []byte("hello world")},
		{name: "test3", src: "dst3", dst: "src3", data: []byte("h")},
		{name: "test4", src: "dst4", dst: "src4", data: []byte("")},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			logger := testLogger()
			src := createTempFile(t, string(tt.data))
			dst := createTempFile(t, "")

			var wg sync.WaitGroup
			wg.Add(1)
			go main.CopyFile(ctx, &wg, src, dst, logger)
			wg.Wait()
			readData, _ := os.ReadFile(dst)
			if !bytes.Equal(readData, tt.data) {
				t.Errorf("Ожидалось, что файл содержит '%s', но он содержит: %s", tt.data, readData)
			}
		})
	}
}

func TestSyncFiles(t *testing.T) {
	cases := []struct {
		name   string
		src    string
		dst    string
		data   []byte
		remove []byte
	}{
		{name: "test1", src: "dst1", dst: "src1", data: []byte{1, 2, 3}, remove: []byte{0}},
		{name: "test2", src: "dst2", dst: "src2", data: []byte("hello world \nword"), remove: []byte("word")},
		{name: "test3", src: "dst3", dst: "src3", data: []byte("h"), remove: []byte("h")},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			logger := testLogger()
			src := createTempFile(t, "")
			dst := createTempFile(t, "")
			var wg sync.WaitGroup
			wg.Add(1)
			go main.writeDatas(ctx, &wg, src, tt.data, logger)
			wg.Wait()
			wg.Add(1)
			go main.syncFiles(ctx, &wg, src, dst, logger)
			wg.Wait()
			readData, _ := os.ReadFile(dst)
			if !bytes.Equal(readData, tt.data) {
				t.Errorf("Ожидалось, что файл содержит '%s', но он содержит: %s", tt.data, readData)
			}
			wg.Add(1)
			go main.removeFiles(ctx, &wg, tt.remove, tt.src, logger)
			wg.Wait()
			readData, _ = os.ReadFile(dst)
			readSrc, _ := os.ReadFile(src)
			if !bytes.Equal(readData, readSrc) {
				t.Errorf("Ожидалось, что файл содержит '%s', но он содержит: %s", tt.data, readData)
			}
		})
	}
}

func TestRemoveFiles(t *testing.T) {
	cases := []struct {
		name   string
		src    string
		data   []byte
		remove []byte
		want   []byte
	}{
		{name: "test1", src: "dst1", data: []byte{1, 2, 3}, remove: []byte{0}, want: []byte{1, 2, 3}},
		{name: "test2", src: "dst2", data: []byte("hello world\n"), remove: []byte("hello world"), want: []byte("")},
		{name: "test3", src: "dst3", data: []byte("h\nw\n"), remove: []byte("h"), want: []byte("w\n")},
		{name: "test4", data: []byte("abcabcabc"), remove: []byte("ab"), want: []byte("ccc")},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			logger := testLogger()
			src := createTempFile(t, "")
			//tt.src = createTempFile(t, "")
			var wg sync.WaitGroup
			wg.Add(1)
			go main.writeDatas(ctx, &wg, src, tt.data, logger)
			wg.Wait()
			wg.Add(1)
			go main.removeFiles(ctx, &wg, tt.remove, src, logger)
			wg.Wait()
			readData, _ := os.ReadFile(src)
			if !bytes.Equal(readData, tt.want) {
				t.Errorf("Ожидалось, что файл содержит '%s', но он содержит: %s", tt.want, readData)
			}
		})
	}
	//rc := createTempFile() и использовать src везде
	//Чисто, безопасно, не нужно придумывать имена
	//Нельзя жёстко задать имя файла
}
func BenchmarkCopyFile(b *testing.B) {
	cases := []struct {
		name string
		data []byte
	}{
		{name: "small", data: []byte("hello world")},
		{name: "medium", data: bytes.Repeat([]byte("abc"), 1024)},            // ~3KB
		{name: "large", data: bytes.Repeat([]byte("1234567890"), 1024*1024)}, // ~10MB
	}

	for _, tt := range cases {
		b.Run(tt.name, func(b *testing.B) {
			ctx := context.Background()
			logger := testLogger()

			// Подготовка исходного файла (один раз, до замера)
			src := createTempFile(b, string(tt.data))

			b.ResetTimer() // Обнуляет таймер перед измерением

			for i := 0; i < b.N; i++ {
				dst := createTempFile(b, "")

				var wg sync.WaitGroup
				wg.Add(1)
				go main.CopyFile(ctx, &wg, src, dst, logger)
				wg.Wait()

				_ = os.Remove(dst) //  Удалим dst после каждой итерации
			}
		})
	}
}
