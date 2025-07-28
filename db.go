package main

import (
	"bytes"
	"context"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

func writeDatas(ctx context.Context, wg *sync.WaitGroup, src string, data []byte, logger *log.Logger) {
	defer wg.Done()

	select {
	case <-ctx.Done():
		return
	default:
	}

	srcFile, err := os.OpenFile(src, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		logger.Printf("[ERROR] Не удалось открыть исходный файл при записи %s: %v", src, err)
	}
	defer srcFile.Close()

	count, err := srcFile.Write(data)
	if err != nil {
		logger.Printf("[ERROR] Не удалось записать данные в исходный файл %s: %v", src, err)
	} else {
		logger.Printf("[INFO] Данные успешно записаны, %d bytes", count)
	}
}

func CopyFile(ctx context.Context, wg *sync.WaitGroup, src, dst string, logger *log.Logger) {
	defer wg.Done()

	select {
	case <-ctx.Done():
		return
	default:
	}

	srcFile, err := os.Open(src)
	if err != nil {
		logger.Printf("[ERROR] Не удалось открыть исходный файл при копировании %s: %v", src, err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		logger.Printf("[ERROR] Не удалось создать файл с БД %s: %v", dst, err)
	}
	defer dstFile.Close()

	count, err := io.Copy(dstFile, srcFile)
	if err != nil {
		logger.Printf("[ERROR] Не удалось скопировать данные в файл с БД %s: %v", dst, err)
	} else {
		logger.Printf("[INFO] Данные успешно скопированы, %d bytes", count)
	}
}
func syncFiles(ctx context.Context, wg *sync.WaitGroup, src, dst string, logger *log.Logger) {
	defer wg.Done()

	select {
	case <-ctx.Done():
		return
	default:
	}

	srcFile, err := os.Open(src)
	if err != nil {
		logger.Printf("[ERROR] При синхронизации не удалось открыть исходный файл %s: %v", src, err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		logger.Printf("[ERROR] Не удалось открыть БД при синхронизации %s: %v", dst, err)
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		logger.Printf("[ERROR] Ошибка при синхронизации : %v", err)
	} else {
		logger.Printf("[INFO] Синхронизация прошла успешно")
	}
}
func removeFiles(ctx context.Context, wg *sync.WaitGroup, forRemove []byte, src string, logger *log.Logger) {
	defer wg.Done()

	select {
	case <-ctx.Done():
		return
	default:
	}

	data, err := os.ReadFile(src)
	if err != nil {
		logger.Printf("[ERROR] Не удалось прочитать файл %s: %v", src, err)
		return
	}

	var result []byte

	if bytes.Contains(data, []byte("\n")) {
		// Построчная фильтрация
		lines := bytes.Split(data, []byte("\n"))
		for _, line := range lines {
			if !bytes.Contains(line, forRemove) {
				result = append(result, line...)
				result = append(result, '\n')
			}
		}
		// Удалим последний \n, если он есть
		result = bytes.TrimSuffix(result, []byte("\n"))
	} else {
		// Фильтрация как подстроку
		result = bytes.ReplaceAll(data, forRemove, []byte{})
	}

	err = os.WriteFile(src, result, 0644)
	if err != nil {
		logger.Printf("[ERROR] Не удалось записать файл %s: %v", src, err)
	} else {
		logger.Printf("[INFO] Успешно удалены вхождения '%s' из файла %s", string(forRemove), src)
	}
}

//	func removeFiles(ctx context.Context, wg *sync.WaitGroup, forRemove []byte, src string, logger *log.Logger) {
//		defer wg.Done()
//
//		select {
//		case <-ctx.Done():
//			return
//		default:
//		}
//
//		srcFile, err := os.Open(src)
//		if err != nil {
//			logger.Printf("[ERROR] При открытии исходного файла произошла ошибка %s: %v", src, err)
//			return
//		}
//		defer srcFile.Close()
//
//		var filteredLines []string
//
//		scanner := bufio.NewScanner(srcFile)
//
//		for scanner.Scan() {
//
//			line := scanner.Bytes()
//			if !strings.Contains(string(line), string(forRemove)) {}
//			//if !bytes.Contains(line, forRemove) {
//				filteredLines = append(filteredLines, string(line))
//			}
//		}
//		if err := scanner.Err(); err != nil {
//			logger.Printf("[ERROR] При сканировании файла произошла ошибка %s: %v", src, err)
//			return
//		}
//
//		// Перезаписываем исходный файл очищенным содержимым
//		err = os.WriteFile(src, []byte(strings.Join(filteredLines, "\n")), 0644)
//		if err != nil {
//			logger.Printf("[ERROR] При записи в файл произошла ошибка %s: %v", src, err)
//		} else {
//			logger.Printf("[INFO] Из файла %s удалены строки, содержащие: %s", src, string(forRemove))
//		}
//	}
func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Настройка логгера в файл
	logFile, err := os.OpenFile("log.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("[ERROR] Не удалось создать лог-файл: %v", err)
	}
	defer logFile.Close()
	logger := log.New(logFile, "", log.LstdFlags)

	src := "src.txt"
	dst := "db.txt"

	// Подготовка исходного файла с тестовыми строками
	_ = os.WriteFile(src, []byte("initial line\nremove this line\nkeep this line\n"), 0644)

	var wg sync.WaitGroup

	// Добавление данных
	dataToWrite := []byte("hello world\n")
	wg.Add(1)
	go writeDatas(ctx, &wg, src, dataToWrite, logger)

	// Копирование
	wg.Add(1)
	go CopyFile(ctx, &wg, src, dst, logger)

	wg.Wait()

	// Удаление строк, содержащих "remove"
	dataToRemove := []byte("remove")
	wg.Add(1)
	go removeFiles(ctx, &wg, dataToRemove, src, logger)

	wg.Wait()

	wg.Add(1)
	go writeDatas(ctx, &wg, src, []byte("\nJohny Seak"), logger)

	// Синхронизация

	for {
		wg.Add(1)
		go syncFiles(ctx, &wg, src, dst, logger)
		time.Sleep(time.Second)
	}
	wg.Wait()

	logger.Println("[INFO] Все операции  завершены")
}
