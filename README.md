# GoldenSky Manager

![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go&logoColor=white)
![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20Windows-lightgrey)
![License](https://img.shields.io/badge/License-TBD-yellow)

Sistema de PDV (Ponto de Venda) para restaurantes, desenvolvido em Go com interface gráfica Fyne.

---

## Overview

GoldenSky Manager is a desktop POS (Point of Sale) application designed for restaurants. It provides a complete order-to-receipt workflow with thermal printer integration, split payment support, kitchen ticket printing, and daily sales analytics.

Built for the Brazilian market with BRL currency formatting, PIX payment support, and a Portuguese-language interface.

**Key highlights:**
- Full order management with notes, discounts, and customer/table tracking
- Thermal printer support via ESC/POS protocol (receipts, kitchen tickets, daily summaries)
- Split payments across cash, card, and PIX
- Daily sales reports with payment breakdown
- Cross-platform: runs on Linux and Windows

## Features

### Order Management
- Create orders with customer name and table number
- Add menu items with per-item notes (e.g., "sem cebola")
- Quantity controls and item removal
- Order-level discounts in BRL
- Automatic order numbering (persistent across sessions)

### Payment Processing
- Three payment methods: Dinheiro (Cash), Cartao (Card), PIX
- Split payments across multiple methods in a single order
- Automatic change calculation for cash payments (handles split scenarios correctly)
- Order finalization with timestamp

### Thermal Printing
- ESC/POS protocol for 80mm thermal printers (tested with GoldenSky GS-T80E)
- **Customer receipts** with restaurant info, itemized list, totals, and payment details
- **Kitchen tickets** with item names and notes only (no prices)
- **Daily summary receipts** with revenue, order count, and payment breakdown
- Cash drawer open command
- CodePage 858 encoding for Portuguese characters (á, é, ç, ã, õ...)
- Auto-detection of connected printers on both Linux and Windows

### Menu Management
- Built-in GUI menu editor (add, edit, remove items)
- CSV bulk import via the `loadmenu` CLI tool
- Category-based organization with tabbed display
- Embedded default menu (75 items across 11 categories) for quick start
- Soft-delete for menu items (deactivate without losing data)

### Sales Analytics
- Daily summary with total revenue, order count, and average ticket value
- Payment method breakdown (cash, card, PIX totals)
- Printable summary receipt for end-of-day closing

### Order History
- Browse past orders by date
- Detailed order view with items, notes, payment method, and timestamps
- Per-date file storage for fast lookup

### Data Persistence
- JSON-based storage (no database required)
- Atomic file writes to prevent corruption
- Per-date order files (`orders_YYYY-MM-DD.json`)
- Backward-compatible schema evolution (old orders load correctly with new fields)
- Thread-safe concurrent writes

## Screenshots

> Screenshots coming soon.

## Requirements

- **Go 1.25+**
- **Fyne v2 system dependencies** (Linux):
  ```
  sudo apt-get install libgl1-mesa-dev xorg-dev
  ```
- **Thermal printer** (optional) — any ESC/POS compatible 80mm printer

## Installation

Clone the repository:

```bash
git clone https://github.com/jixkls/goldensky-manager.git
cd goldensky-manager
```

Install Go dependencies:

```bash
go mod download
```

Build for Linux:

```bash
go build -o goldensky-pos .
```

Cross-compile for Windows:

```bash
GOOS=windows GOARCH=amd64 go build -o goldensky-pos.exe .
```

## Usage

### Running the Application

```bash
./goldensky-pos
```

The application opens a 1280x768 window with three panels: menu categories (left), current order (center), and actions (right).

### First-Run Configuration

On first launch, GoldenSky creates a default configuration. Open the settings dialog to configure:

1. **Restaurant info** — name, address, phone, CNPJ, receipt footer message
2. **Printer path** — device path (Linux) or printer name (Windows)
3. **Kitchen ticket toggle** — enable/disable kitchen ticket printing

### Menu Setup

**Option A: GUI Editor**
Use the built-in menu editor to add items one by one with name, price, and category.

**Option B: CSV Import**
Prepare a CSV file and use the `loadmenu` tool for bulk import (see [CLI Tools](#cli-tools)).

### Order Workflow

1. Select items from the category tabs on the left panel
2. Adjust quantities and add notes as needed
3. Set customer name and table number (optional)
4. Apply discount if applicable
5. Choose payment method (or split across multiple methods)
6. Finalize the order — receipt prints automatically if a printer is connected

## CLI Tools

### `loadmenu` — CSV Menu Import

Imports menu items from a CSV file into GoldenSky's menu.

**Build:**

```bash
go build -o loadmenu ./cmd/loadmenu
```

**Usage:**

```bash
./loadmenu cardapio.csv
```

**CSV format:**

| Column | Field | Required |
|--------|-------|----------|
| 1 | Categoria (category) | Yes |
| 2 | Item (name) | Yes |
| 3 | Descricao (description) | No (ignored) |
| 4 | Peso/Volume (weight/size) | No (ignored) |
| 5 | Preco R$ (price in BRL) | Yes |

**Example CSV:**

```csv
Categoria,Item,Descricao,Peso/Volume,Preco (R$)
Pizzas Tradicionais,Portuguesa,"Molho, Presunto, Mussarela",,53.00
Cervejas - Lata,Skol Lata,,350ml,5.00
Refrigerantes,Coca-Cola,,1L,8.00
```

Prices accept both dot (`53.00`) and comma (`53,00`) as decimal separator.

## Project Structure

```
goldensky-manager/
├── main.go                        # Application entry point
├── go.mod                         # Module definition (Go 1.25, Fyne v2)
│
├── cmd/
│   └── loadmenu/
│       └── main.go                # CSV menu import CLI tool
│
├── internal/
│   ├── pos/                       # Domain logic
│   │   ├── order.go               # Order, menu, payment models and operations
│   │   ├── order_test.go          # Core functionality tests
│   │   └── order_compat_test.go   # Backward compatibility tests
│   │
│   ├── printer/                   # Thermal printer integration
│   │   ├── escpos.go              # ESC/POS command constants and receipt builder
│   │   ├── connection.go          # Printer connection interface
│   │   ├── connection_linux.go    # Linux USB device connection
│   │   ├── connection_windows.go  # Windows Spooler API connection
│   │   ├── receipt.go             # Customer receipt formatting
│   │   ├── kitchen_ticket.go      # Kitchen ticket formatting (no prices)
│   │   └── summary_receipt.go     # Daily summary receipt
│   │
│   └── storage/                   # Data persistence
│       ├── config.go              # Configuration load/save
│       ├── orders.go              # Order persistence (JSON, per-date files)
│       ├── default_menu.json      # Embedded default menu (75 items)
│       ├── defaults_linux.go      # Linux default paths
│       └── defaults_windows.go    # Windows default paths
│
├── ui/                            # Fyne GUI
│   ├── gui.go                     # App initialization and layout
│   ├── menu_panel.go              # Category tabs and item buttons
│   ├── order_panel.go             # Current order display and editing
│   ├── action_panel.go            # Payment and order finalization
│   ├── status_bar.go              # Printer connection status
│   ├── dialogs.go                 # Settings and menu editor dialogs
│   ├── history_dialog.go          # Order history browser
│   ├── summary_dialog.go          # Daily sales summary view
│   └── icon.go                    # App icon resource
│
└── winres/                        # Windows build resources
    ├── winres.json
    └── icon256.png
```

## Configuration

Configuration is stored as JSON at:

| Platform | Path |
|----------|------|
| Linux | `~/.config/goldensky-pos/config.json` |
| Windows | `%APPDATA%\goldensky-pos\config.json` |

**Config fields:**

```json
{
  "restaurant": {
    "name": "Meu Restaurante",
    "address": "Rua Exemplo, 123",
    "phone": "(11) 9999-9999",
    "cnpj": "12.345.678/0001-99",
    "footer": "Obrigado pela preferencia!"
  },
  "printer": {
    "device_path": "/dev/usb/lp0",
    "chars_per_line": 48
  },
  "order_counter": 0,
  "kitchen_ticket": false
}
```

Menu data is stored alongside the config as `menu.json`. Orders are stored in the `orders/` subdirectory with one file per date (`orders_YYYY-MM-DD.json`).

## Printer Setup

### Linux

Default device path: `/dev/usb/lp0`

Add your user to the `lp` group for access without root:

```bash
sudo usermod -aG lp $USER
```

Log out and back in for the group change to take effect. The application auto-detects printers by scanning `/dev/usb/lp*`.

### Windows

The application uses the Windows Spooler API to communicate with installed printers. Set the printer name in the configuration dialog (leave empty for auto-detection via `EnumPrintersW`).

### Supported ESC/POS Commands

| Command | Description |
|---------|-------------|
| `ESC @` | Printer initialization/reset |
| `ESC a` | Text alignment (left, center, right) |
| `ESC E` | Bold on/off |
| `ESC !` | Font size (normal, double height+width) |
| `GS V` | Paper cut (partial and full) |
| `ESC p` | Cash drawer open |
| `ESC t` | CodePage 858 selection (Portuguese charset) |
| `ESC d` | Feed N lines |

## Testing

Run all tests:

```bash
go test ./...
```

There are **22 tests** covering:

- BRL currency formatting (`R$ 1.234,56`)
- Order totals, subtotals, and discounts
- Menu category filtering
- Single and split payment finalization
- Cash change calculation
- Daily summary computation with payment breakdown
- Backward compatibility with pre-split-payment order format
- JSON serialization round-trips

## Tech Stack

| Component | Technology |
|-----------|------------|
| Language | Go 1.25 |
| GUI Framework | [Fyne v2](https://fyne.io/) |
| Printer Protocol | ESC/POS |
| Data Storage | JSON files |
| Text Encoding | CodePage 858 via `golang.org/x/text` |

## License

TBD
