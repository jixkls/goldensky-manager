package ui

import (
	"fmt"
	"log"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"notinha/internal/pos"
	"notinha/internal/printer"
	"notinha/internal/storage"
)

// App holds the entire application state and UI references.
type App struct {
	fyneApp    fyne.App
	mainWindow fyne.Window
	config     *storage.Config
	menu       *pos.Menu
	order      *pos.Order
	printer    *printer.Printer

	// Split payment state
	splitPayments []pos.PaymentSplit

	// UI widget references
	orderList        *widget.List
	totalLabel       *widget.Label
	customerEntry    *widget.Entry
	tableEntry       *widget.Entry
	discountEntry    *widget.Entry
	paymentRadio     *widget.RadioGroup
	kitchenCheck     *widget.Check
	cashReceivedEntry *widget.Entry
	changeLabel      *widget.Label
	cashSection      *fyne.Container
	statusLabel      *widget.Label
	menuTabs         *container.AppTabs
}

// NewApp creates and initializes the application.
func NewApp() *App {
	a := &App{}

	cfg, err := storage.LoadConfig()
	if err != nil {
		log.Printf("Aviso: erro ao carregar config: %v", err)
	}
	a.config = cfg

	menu, err := storage.LoadMenu()
	if err != nil {
		log.Printf("Aviso: erro ao carregar cardapio: %v", err)
	}
	a.menu = menu

	a.order = pos.NewOrder(cfg.NextOrderNumber())

	a.fyneApp = app.New()
	a.fyneApp.SetIcon(appIcon)
	a.mainWindow = a.fyneApp.NewWindow("GoldenSky POS")
	a.mainWindow.Resize(fyne.NewSize(1280, 768))

	a.connectPrinter()
	a.buildLayout()

	return a
}

// Run starts the application event loop.
func (a *App) Run() {
	a.mainWindow.ShowAndRun()
}

func (a *App) connectPrinter() {
	p, err := printer.Open(a.config.Printer.DevicePath)
	if err == nil {
		a.printer = p
		return
	}
	log.Printf("Impressora nao encontrada em %s: %v", a.config.Printer.DevicePath, err)

	for _, path := range printer.DetectPrinters() {
		p, err = printer.Open(path)
		if err == nil {
			a.printer = p
			a.config.Printer.DevicePath = path
			_ = storage.SaveConfig(a.config)
			log.Printf("Impressora detectada em %s", path)
			return
		}
	}
	log.Println("Nenhuma impressora detectada")
}

func (a *App) buildLayout() {
	menuPanel := a.buildMenuPanel()
	orderPanel := a.buildOrderPanel()
	actionPanel := a.buildActionPanel()
	statusBar := a.buildStatusBar()

	leftPanel := container.NewStack(menuPanel)
	rightPanel := container.NewStack(actionPanel)

	// Use Border layout: left=menu, right=actions, center=order, bottom=status
	centerContent := container.NewBorder(
		nil,       // top
		statusBar, // bottom
		leftPanel, // left
		rightPanel, // right
		orderPanel, // center
	)

	a.mainWindow.SetContent(centerContent)

	toolbar := a.buildToolbar()
	a.mainWindow.SetMainMenu(toolbar)
}

func (a *App) buildToolbar() *fyne.MainMenu {
	configItem := fyne.NewMenuItem("Configuracoes", func() {
		a.showConfigDialog()
	})
	menuEditorItem := fyne.NewMenuItem("Editar Cardapio", func() {
		a.showMenuEditorDialog()
	})
	historyItem := fyne.NewMenuItem("Historico de Pedidos", func() {
		a.showOrderHistoryDialog()
	})
	summaryItem := fyne.NewMenuItem("Resumo do Dia", func() {
		a.showDaySummaryDialog()
	})
	settingsMenu := fyne.NewMenu("Opcoes", configItem, menuEditorItem,
		fyne.NewMenuItemSeparator(), historyItem, summaryItem)
	return fyne.NewMainMenu(settingsMenu)
}

func (a *App) newOrder() {
	a.order = pos.NewOrder(a.config.NextOrderNumber())
	a.splitPayments = nil
	a.customerEntry.SetText("")
	a.tableEntry.SetText("")
	a.discountEntry.SetText("")
	a.paymentRadio.SetSelected(string(pos.PaymentDinheiro))
	a.paymentRadio.Enable()
	if a.cashReceivedEntry != nil {
		a.cashReceivedEntry.SetText("")
	}
	if a.changeLabel != nil {
		a.changeLabel.SetText("")
	}
	if a.cashSection != nil {
		a.cashSection.Show()
	}
	a.refreshOrderDisplay()
}

func (a *App) addItemToOrder(item pos.MenuItem) {
	notesEntry := widget.NewMultiLineEntry()
	notesEntry.SetPlaceHolder("Observacoes (opcional)")
	notesEntry.SetMinRowsVisible(3)

	content := container.NewVBox(
		widget.NewLabel(fmt.Sprintf("%s - %s", item.Name, pos.FormatBRL(item.Price))),
		notesEntry,
	)

	d := dialog.NewCustomConfirm("Adicionar Item", "Adicionar", "Cancelar", content, func(ok bool) {
		if !ok {
			return
		}
		a.order.AddItem(item, 1, strings.TrimSpace(notesEntry.Text))
		a.refreshOrderDisplay()
	}, a.mainWindow)
	d.Show()
}

func (a *App) showEditNotesDialog(index int) {
	if index < 0 || index >= len(a.order.Items) {
		return
	}
	orderItem := a.order.Items[index]

	notesEntry := widget.NewMultiLineEntry()
	notesEntry.SetText(orderItem.Notes)
	notesEntry.SetMinRowsVisible(3)

	content := container.NewVBox(
		widget.NewLabel(orderItem.Item.Name),
		notesEntry,
	)

	d := dialog.NewCustomConfirm("Editar Observacao", "Salvar", "Cancelar", content, func(ok bool) {
		if !ok {
			return
		}
		a.order.UpdateNotes(index, strings.TrimSpace(notesEntry.Text))
		a.refreshOrderDisplay()
	}, a.mainWindow)
	d.Show()
}

func (a *App) refreshOrderDisplay() {
	a.orderList.Refresh()
	a.totalLabel.SetText(pos.FormatBRL(a.order.Total()))
}
