package storage

import (
	_ "embed"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"notinha/internal/pos"
)

//go:embed default_menu.json
var defaultMenuJSON []byte

type RestaurantInfo struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Phone   string `json:"phone"`
	CNPJ    string `json:"cnpj"`
	Footer  string `json:"footer"`
}

type PrinterConfig struct {
	DevicePath   string `json:"device_path"`
	CharsPerLine int    `json:"chars_per_line"`
}

type Config struct {
	Restaurant    RestaurantInfo `json:"restaurant"`
	Printer       PrinterConfig  `json:"printer"`
	OrderCounter  int            `json:"order_counter"`
	KitchenTicket bool           `json:"kitchen_ticket"`

	mu sync.Mutex
}

func DefaultConfig() *Config {
	return &Config{
		Restaurant: RestaurantInfo{
			Name:   "Meu Restaurante",
			Footer: "Obrigado pela preferencia!",
		},
		Printer: PrinterConfig{
			DevicePath:   defaultPrinterPath,
			CharsPerLine: 48,
		},
		OrderCounter: 0,
	}
}

func configDir() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	path := filepath.Join(dir, "goldensky-pos")
	return path, os.MkdirAll(path, 0755)
}

func configPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

func menuPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "menu.json"), nil
}

func LoadConfig() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return DefaultConfig(), err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			cfg := DefaultConfig()
			_ = SaveConfig(cfg)
			return cfg, nil
		}
		return DefaultConfig(), err
	}

	cfg := DefaultConfig()
	if err := json.Unmarshal(data, cfg); err != nil {
		return DefaultConfig(), err
	}
	return cfg, nil
}

func SaveConfig(cfg *Config) error {
	path, err := configPath()
	if err != nil {
		return err
	}
	return atomicWriteJSON(path, cfg)
}

func LoadMenu() (*pos.Menu, error) {
	path, err := menuPath()
	if err != nil {
		return pos.NewMenu(), err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			menu := pos.NewMenu()
			_ = json.Unmarshal(defaultMenuJSON, menu)
			_ = SaveMenu(menu)
			return menu, nil
		}
		return pos.NewMenu(), err
	}

	menu := pos.NewMenu()
	if err := json.Unmarshal(data, menu); err != nil {
		return pos.NewMenu(), err
	}
	return menu, nil
}

func SaveMenu(menu *pos.Menu) error {
	path, err := menuPath()
	if err != nil {
		return err
	}
	return atomicWriteJSON(path, menu)
}

func (c *Config) NextOrderNumber() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.OrderCounter++
	_ = SaveConfig(c)
	return c.OrderCounter
}

func atomicWriteJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}
