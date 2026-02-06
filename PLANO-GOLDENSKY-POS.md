# GoldenSky POS - Plano do Projeto

Sistema de ponto de venda para restaurante com impressora térmica GoldenSky GS-T80E.

---

## Stack Tecnológica

| Componente | Tecnologia |
|------------|------------|
| Linguagem | Go 1.21+ |
| Interface Gráfica | Fyne v2 |
| Conexão Impressora | USB direto (file I/O) |
| Protocolo | ESC/POS |
| Persistência | JSON (arquivos locais) |
| Idioma | Português (Brasil) |

---

## Especificações da Impressora

| Spec | Valor |
|------|-------|
| Modelo | GS-T80E |
| Marca | GoldenSky |
| Papel | 80mm (48 caracteres por linha) |
| Velocidade | 200mm/s |
| Interface | USB + LAN |
| Comandos | ESC/POS |
| Gaveta | 24V, 1A |
| Serial | GST80E-BLU2405080076 |

---

## Estrutura de Diretórios

```
goldensky-pos/
├── main.go                      # Entrada da aplicação
├── go.mod                       # Módulo Go e dependências
├── go.sum                       # Checksums das dependências
│
├── internal/
│   ├── printer/
│   │   ├── escpos.go            # Comandos ESC/POS (bytes)
│   │   ├── connection.go        # Conexão USB com a impressora
│   │   └── receipt.go           # Formatação de cupons
│   │
│   ├── pos/
│   │   └── order.go             # Pedidos, itens, cardápio, cálculos
│   │
│   └── storage/
│       └── config.go            # Configurações (JSON)
│
├── ui/
│   └── gui.go                   # Interface Fyne
│
└── assets/
    └── icon.ico                 # Ícone do app (opcional)
```

---

## Arquivos e Responsabilidades

### `main.go`
Ponto de entrada. Inicializa a aplicação e executa o loop principal do Fyne.

### `internal/printer/escpos.go`
Define os comandos ESC/POS como slices de bytes:
- Inicialização (`0x1B 0x40`)
- Alinhamento (esquerda, centro, direita)
- Negrito, sublinhado
- Tamanho de fonte (normal, duplo, triplo)
- Corte de papel (parcial/total)
- Abertura de gaveta
- Code page para português (PC858)

Implementa `ReceiptBuilder` com interface fluente para construir cupons.

### `internal/printer/connection.go`
Gerencia conexão USB:
- Abre dispositivo (`/dev/usb/lp0` no Linux, `COM3` no Windows)
- Envia bytes para impressora
- Detecta impressoras disponíveis
- Trata erros de conexão

### `internal/printer/receipt.go`
Formata cupons de venda:
- Cabeçalho (nome, endereço, telefone, CNPJ)
- Data/hora
- Número do pedido
- Lista de itens (quantidade, nome, preço)
- Subtotal, desconto, total
- Mensagem de rodapé
- Formatação de valores em Real (R$ 1.234,56)

### `internal/pos/order.go`
Estruturas de dados:
- `MenuItem`: id, nome, preço, categoria, ativo
- `OrderItem`: item + quantidade + observações
- `Order`: itens, cliente, mesa, desconto, pagamento, status
- `Menu`: lista de itens com CRUD

Métodos:
- Adicionar/remover itens
- Calcular subtotal/total
- Salvar/carregar cardápio (JSON)

### `internal/storage/config.go`
Configurações persistentes:
- Dados do restaurante
- Caminho da impressora
- Preferências de impressão
- Contador de pedidos

Armazenamento por OS:
- Windows: `%APPDATA%\goldensky-pos\`
- Linux: `~/.config/goldensky-pos/`
- macOS: `~/Library/Application Support/goldensky-pos/`

### `ui/gui.go`
Interface gráfica com Fyne:
- Painel esquerdo: cardápio em abas por categoria
- Painel central: lista do pedido atual
- Painel direito: finalização e ações
- Barra de status: conexão da impressora

---

## Dependências

```go
// go.mod
module goldensky-pos

go 1.21

require (
    fyne.io/fyne/v2 v2.4.4
)
```

---

## Comandos ESC/POS Necessários

| Comando | Bytes | Função |
|---------|-------|--------|
| Init | `1B 40` | Reset da impressora |
| Align Left | `1B 61 00` | Alinhar esquerda |
| Align Center | `1B 61 01` | Alinhar centro |
| Align Right | `1B 61 02` | Alinhar direita |
| Bold On | `1B 45 01` | Ativar negrito |
| Bold Off | `1B 45 00` | Desativar negrito |
| Font Normal | `1D 21 00` | Tamanho normal |
| Font Double | `1D 21 11` | Tamanho duplo |
| Line Feed | `0A` | Nova linha |
| Feed N Lines | `1B 64 N` | Avançar N linhas |
| Partial Cut | `1D 56 01` | Corte parcial |
| Full Cut | `1D 56 00` | Corte total |
| Open Drawer | `1B 70 00 19 FA` | Abrir gaveta |
| Code Page 858 | `1B 74 13` | Charset português |

---

## Funcionalidades v1.0

- [x] Cardápio configurável por categoria
- [x] Adicionar/remover itens do pedido
- [x] Cálculo automático de totais
- [x] Impressão de cupom formatado
- [x] Nome do cliente (opcional)
- [x] Número da mesa (opcional)
- [x] Formas de pagamento (Dinheiro, Cartão, Pix, etc.)
- [x] Teste de impressão
- [x] Abertura de gaveta
- [x] Configurações do restaurante
- [x] Editor de cardápio
- [x] Persistência de dados (JSON)

---

## Funcionalidades Futuras (v2.0)

- [ ] Cálculo de impostos
- [ ] Cupom de cozinha (formato simplificado)
- [ ] Histórico de pedidos
- [ ] Resumo diário de vendas
- [ ] Conexão via rede (LAN)
- [ ] Backup automático

---

## Compilação

```bash
# Instalar dependências
go mod tidy

# Linux
go build -o goldensky-pos

# Windows
go build -o goldensky-pos.exe

# Windows (a partir do Linux)
GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc go build -o goldensky-pos.exe
```

---

## Requisitos de Sistema

### Windows
- Go 1.21+
- TDM-GCC ou MinGW-w64 (para Fyne)

### Linux
```bash
sudo apt-get install golang-go gcc libgl1-mesa-dev xorg-dev
sudo usermod -a -G lp $USER  # Permissão para impressora
```

---

## Caminhos da Impressora

| Sistema | Caminho Típico |
|---------|----------------|
| Linux | `/dev/usb/lp0` |
| Windows | `COM3` ou `COM4` |

---

## Dados do Restaurante (a obter)

- Nome do restaurante
- Endereço
- Telefone
- CNPJ
- Lista completa de itens do cardápio com preços e categorias
