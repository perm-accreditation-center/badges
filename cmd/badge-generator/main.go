package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"badges/internal/aggregate"
	"badges/internal/docx"
	"badges/internal/input"
	"badges/internal/model"
	"badges/internal/ooxml"
	"badges/internal/spo"
)

func main() {
	root, err := os.Executable()
	if err == nil {
		root = filepath.Dir(root)
	} else {
		root, _ = os.Getwd()
	}
	in, out := filepath.Join(root, "input"), filepath.Join(root, "output")
	_ = os.MkdirAll(in, 0755)
	_ = os.MkdirAll(out, 0755)
	fmt.Println("Генератор бейджей для аккредитации — профиль СПО")
	fmt.Println("Исходные DOCX: input | Результаты: output | Карточка: 90 × 55 мм")
	files, diags := input.Discover(in)
	fmt.Printf("[1/5] Папки input/output: готово\n[2/5] Найдено файлов СПО: %d\n", len(files))
	var records []model.Record
	for _, file := range files {
		document, err := ooxml.ReadDocument(file)
		if err != nil {
			fmt.Println("Пропущен", filepath.Base(file), "—", err)
			continue
		}
		r, d := spo.Parse(document, file)
		records = append(records, r...)
		diags = append(diags, d...)
	}
	people, d := aggregate.Records(records)
	diags = append(diags, d...)
	fmt.Printf("[3/5] Извлечено записей: %d\n[4/5] Уникальных людей: %d\n", len(records), len(people))
	if len(people) == 0 {
		fmt.Println("Нет корректных записей в input.")
		pause()
		return
	}
	bundle := filepath.Join(out, time.Now().Format("2006-01-02_150405"))
	_ = os.MkdirAll(bundle, 0755)
	path := filepath.Join(bundle, "бейджи.docx")
	if err := docx.Write(path, people); err != nil {
		fmt.Println("Ошибка создания DOCX:", err)
		pause()
		return
	}
	fmt.Printf("[5/5] Word-файл создан: %s\nПредупреждений: %d\n\nПечать: «Фактический размер / 100%%», не «Подогнать».\n", path, len(diags))
	pause()
}

func pause() {
	fmt.Println("Нажмите Enter, чтобы закрыть окно...")
	_, _ = bufio.NewReader(os.Stdin).ReadString('\n')
}
