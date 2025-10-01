package ui

import (
	"image/color"
	"log"
	"os/exec"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/probeldev/fastlauncher/model"
	"github.com/probeldev/fastlauncher/pkg/apprunner"
)

type item struct {
	title   string
	command string
}

type uiModel struct {
	items       []item
	commands    map[string]string
	filtered    []string
	input       *widget.Entry
	list        *widget.List
	currentItem int
	window      fyne.Window
}

// filterItems фильтрует элементы по запросу (fuzzy search как в TUI версии)
func (m *uiModel) filterItems(query string) []string {
	if query == "" {
		// Возвращаем все ключи
		allKeys := make([]string, 0, len(m.commands))
		for key := range m.commands {
			allKeys = append(allKeys, key)
		}
		return allKeys
	}

	query = strings.ToLower(query)
	var filtered []string

	for key := range m.commands {
		title := strings.ToLower(key)
		if fuzzyMatch(title, query) {
			filtered = append(filtered, key)
		}
	}

	return filtered
}

// fuzzyMatch проверяет, можно ли найти query как подпоследовательность в str
func fuzzyMatch(str, query string) bool {
	if query == "" {
		return true
	}
	if str == "" {
		return false
	}

	// Ищем первую букву запроса в строке
	firstChar := query[0]
	pos := strings.IndexByte(str, firstChar)
	if pos == -1 {
		return false
	}

	// Рекурсивно проверяем оставшуюся часть запроса
	return fuzzyMatch(str[pos+1:], query[1:])
}

// updateList обновляет содержимое списка
func (m *uiModel) updateList() {
	m.filtered = m.filterItems(m.input.Text)

	// Сбрасываем текущий элемент при обновлении списка
	if len(m.filtered) > 0 {
		m.currentItem = 0
	} else {
		m.currentItem = -1
	}

	if m.list != nil {
		m.list.Refresh()
		if m.currentItem >= 0 {
			m.list.Select(m.currentItem)
		}
	}
}

// executeCommand выполняет команду (аналогично TUI версии)
func (m *uiModel) executeCommand(cmd string) {
	log.Printf("Executing: %s", cmd)

	// Специальная обработка для Ghostty (из Fyne версии)
	if strings.HasPrefix(cmd, "open -a Ghostty") {
		parts := strings.Fields(cmd)
		if len(parts) >= 4 {
			path := strings.Join(parts[3:], " ")
			m.openGhostty(path)
			return
		}
	}

	// Используем apprunner из TUI версии
	runner, err := apprunner.GetAppRunner(apprunner.OsLinux) // или определите ОС динамически
	if err != nil {
		log.Println("GetAppRunner error:", err)
		return
	}

	err = runner.Run(cmd)
	if err != nil {
		log.Println("Run error:", err)
		return
	}
}

func (m *uiModel) openGhostty(path string) {
	expandedPath, _ := filepath.Abs(path)
	cmd := exec.Command("open", "-a", "Ghostty", expandedPath)
	go cmd.Run()
}

// moveSelection перемещает выделение вверх или вниз
func (m *uiModel) moveSelection(direction int) {
	if len(m.filtered) == 0 {
		return
	}

	newIndex := m.currentItem + direction
	if newIndex < 0 {
		newIndex = 0
	} else if newIndex >= len(m.filtered) {
		newIndex = len(m.filtered) - 1
	}

	m.currentItem = newIndex
	m.list.Select(m.currentItem)
	m.list.Refresh()
}

// executeSelected выполняет выбранную команду
func (m *uiModel) executeSelected() {
	if m.currentItem >= 0 && m.currentItem < len(m.filtered) {
		selectedKey := m.filtered[m.currentItem]
		if cmd, exists := m.commands[selectedKey]; exists {
			m.executeCommand(cmd)
			m.window.Close()
		}
	}
}

// CustomListItem создает кастомный элемент списка с выделением
type CustomListItem struct {
	widget.BaseWidget
	title       *canvas.Text
	description *canvas.Text
	background  *canvas.Rectangle
	isSelected  bool
}

// NewCustomListItem создает новый элемент списка
func NewCustomListItem(title, description string, isSelected bool) *CustomListItem {
	item := &CustomListItem{
		title:       canvas.NewText(title, color.Black),
		description: canvas.NewText(description, color.Gray{0x80}),
		background:  canvas.NewRectangle(color.White),
		isSelected:  isSelected,
	}

	item.title.TextStyle = fyne.TextStyle{Bold: isSelected}
	item.title.TextSize = 14
	item.description.TextSize = 12

	if isSelected {
		item.background.FillColor = color.NRGBA{R: 0x33, G: 0x99, B: 0xff, A: 0x99} // Голубой с прозрачностью
	} else {
		item.background.FillColor = color.White
	}

	item.ExtendBaseWidget(item)
	return item
}

// CreateRenderer создает рендерер для элемента
func (i *CustomListItem) CreateRenderer() fyne.WidgetRenderer {
	content := container.NewVBox(
		container.NewHBox(i.title),
		container.NewHBox(i.description),
	)

	paddedContent := container.NewPadded(content)
	fullContent := container.NewStack(i.background, paddedContent)

	return widget.NewSimpleRenderer(fullContent)
}

// createCustomList создает кастомный список с выделением
func (m *uiModel) createCustomList() *widget.List {
	list := widget.NewList(
		func() int {
			return len(m.filtered)
		},
		func() fyne.CanvasObject {
			// Создаем элемент с дефолтными значениями
			return NewCustomListItem("template", "description", false)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			if i < len(m.filtered) {
				key := m.filtered[i]
				item := o.(*CustomListItem)

				// Обновляем текст
				item.title.Text = key
				item.description.Text = m.commands[key]

				// Обновляем выделение
				item.isSelected = (i == m.currentItem)
				if item.isSelected {
					item.background.FillColor = color.NRGBA{R: 0x33, G: 0x99, B: 0xff, A: 0x99}
					item.title.TextStyle = fyne.TextStyle{Bold: true}
					item.title.Color = color.White
					item.description.Color = color.White
				} else {
					item.background.FillColor = color.White
					item.title.TextStyle = fyne.TextStyle{}
					item.title.Color = color.Black
					item.description.Color = color.Gray{0x80}
				}

				item.Refresh()
			}
		},
	)

	// Обработка выбора из списка
	list.OnSelected = func(id widget.ListItemID) {
		m.currentItem = id
		m.executeSelected()
	}

	return list
}

// handleKeyPress обрабатывает нажатия клавиш
func (m *uiModel) handleKeyPress(key *fyne.KeyEvent) {
	switch key.Name {
	case fyne.KeyUp:
		m.moveSelection(-1)
	case fyne.KeyDown:
		m.moveSelection(1)
	case fyne.KeyReturn, fyne.KeyEnter:
		m.executeSelected()
	case fyne.KeyEscape:
		m.window.Close()
	}
}

func StartUI(apps []model.App) {
	myApp := app.New()
	myWindow := myApp.NewWindow("Fast Launcher")
	myWindow.SetFixedSize(true)
	myWindow.Resize(fyne.NewSize(600, 400))
	myWindow.CenterOnScreen()

	// Создаём модель как в TUI версии
	m := &uiModel{
		items:    make([]item, len(apps)),
		commands: make(map[string]string),
		window:   myWindow,
	}

	// Заполняем элементы как в TUI версии и создаем commands map для Fyne
	for i, a := range apps {
		m.items[i] = item{
			title:   a.Title,
			command: a.Command,
		}
		m.commands[a.Title] = a.Command
	}

	// Поле ввода как в Fyne версии, но с логикой из TUI
	input := widget.NewEntry()
	input.SetPlaceHolder("Введите команду для поиска...")
	m.input = input

	// Создаем кастомный список
	list := m.createCustomList()
	m.list = list

	// Обработка ввода - используем fuzzy search из TUI версии
	input.OnChanged = func(text string) {
		m.updateList()
	}

	// Обработка Enter - комбинация обеих версий
	input.OnSubmitted = func(cmd string) {
		// Если есть выбранный элемент, выполняем его
		if m.currentItem >= 0 && m.currentItem < len(m.filtered) {
			m.executeSelected()
			return
		}

		// Иначе пытаемся найти точное совпадение
		if command, exists := m.commands[cmd]; exists {
			m.executeCommand(command)
			myWindow.Close()
			return
		}

		// Если ничего не найдено, выполняем как есть
		m.executeCommand(cmd)
		myWindow.Close()
	}

	// Обработка клавиш для навигации стрелками на уровне окна
	myWindow.Canvas().SetOnTypedKey(func(key *fyne.KeyEvent) {
		// Если фокус на поле ввода или списке, обрабатываем навигацию
		m.handleKeyPress(key)
	})

	// Компоновка как в Fyne версии
	content := container.NewBorder(
		input, // сверху - поле ввода
		nil, nil, nil,
		list, // по центру - список подсказок
	)

	myWindow.SetContent(content)

	// Фокус на поле ввода при открытии
	myWindow.Canvas().Focus(input)

	// Инициализируем список
	m.updateList()

	myWindow.ShowAndRun()
}
