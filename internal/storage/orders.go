package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"notinha/internal/pos"
)

var ordersMu sync.Mutex

func ordersDir() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	path := filepath.Join(dir, "orders")
	return path, os.MkdirAll(path, 0755)
}

func ordersFilePath(date string) (string, error) {
	dir, err := ordersDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "orders_"+date+".json"), nil
}

func SaveOrder(order *pos.Order) error {
	ordersMu.Lock()
	defer ordersMu.Unlock()

	date := order.ClosedAt.Format("2006-01-02")
	path, err := ordersFilePath(date)
	if err != nil {
		return err
	}

	orders, err := loadOrdersFromFile(path)
	if err != nil {
		return err
	}

	orders = append(orders, *order)
	return atomicWriteJSON(path, orders)
}

func LoadDayOrders(date string) ([]pos.Order, error) {
	ordersMu.Lock()
	defer ordersMu.Unlock()

	path, err := ordersFilePath(date)
	if err != nil {
		return nil, err
	}
	return loadOrdersFromFile(path)
}

func ListOrderDates() ([]string, error) {
	dir, err := ordersDir()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var dates []string
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, "orders_") && strings.HasSuffix(name, ".json") {
			date := strings.TrimPrefix(name, "orders_")
			date = strings.TrimSuffix(date, ".json")
			dates = append(dates, date)
		}
	}

	sort.Sort(sort.Reverse(sort.StringSlice(dates)))
	return dates, nil
}

func TodayDateString() string {
	return time.Now().Format("2006-01-02")
}

func loadOrdersFromFile(path string) ([]pos.Order, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var orders []pos.Order
	if err := json.Unmarshal(data, &orders); err != nil {
		return nil, err
	}
	return orders, nil
}
