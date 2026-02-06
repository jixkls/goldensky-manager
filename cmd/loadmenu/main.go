package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"notinha/internal/pos"
	"notinha/internal/storage"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Uso: loadmenu <arquivo.csv>")
	}

	f, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatalf("Erro ao abrir CSV: %v", err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.LazyQuotes = true
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("Erro ao ler CSV: %v", err)
	}

	menu := pos.NewMenu()

	for i, row := range records {
		if i == 0 {
			continue // skip header
		}
		if len(row) < 5 {
			continue
		}

		category := strings.TrimSpace(row[0])
		name := strings.TrimSpace(row[1])
		priceStr := strings.TrimSpace(row[4])

		if category == "" || name == "" || priceStr == "" {
			continue
		}

		priceStr = strings.ReplaceAll(priceStr, ",", ".")
		price, err := strconv.ParseFloat(priceStr, 64)
		if err != nil {
			log.Printf("Linha %d: preco invalido %q, ignorando", i+1, priceStr)
			continue
		}
		centavos := int64(price * 100)

		menu.AddItem(pos.MenuItem{
			Name:     name,
			Price:    centavos,
			Category: category,
		})
	}

	if err := storage.SaveMenu(menu); err != nil {
		log.Fatalf("Erro ao salvar cardapio: %v", err)
	}

	fmt.Printf("Cardapio salvo com %d itens!\n", len(menu.Items))
	for _, cat := range menu.Categories() {
		items := menu.ItemsByCategory(cat)
		fmt.Printf("  %s: %d itens\n", cat, len(items))
	}
}
