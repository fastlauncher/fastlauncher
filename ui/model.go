// Package ui - implementation ui
package ui

import (
	"image/color"
	"log"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/probeldev/fastlauncher/model"
	"github.com/probeldev/fastlauncher/pkg/apprunner"
)

type uiModel struct {
	items           []model.App
	filtered        []model.App
	input           *widget.Entry
	list            *widget.List
	currentItem     int
	window          fyne.Window
	ignoreSelection bool // Флаг для игнорирования события выбора
}

// filterItems фильтрует элементы по запросу (fuzzy search как в TUI версии)
func (m *uiModel) filterItems(query string) []model.App {
	if query == "" {
		return m.items
	}

	query = strings.ToLower(query)
	var filtered []model.App

	for _, item := range m.items {
		title := strings.ToLower(item.Title)
		if fuzzyMatch(title, query) {
			filtered = append(filtered, item)
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
			m.ignoreSelection = true
			m.list.Select(m.currentItem)
			m.ignoreSelection = false
		}
	}
}

// executeCommand выполняет команду (аналогично TUI версии)
func (m *uiModel) executeCommand(cmd string) {

	log.Println("command:", cmd)

	// Используем apprunner из TUI версии
	// TODO: GET OS
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
	m.ignoreSelection = true
	m.list.Select(m.currentItem)
	m.ignoreSelection = false
	m.list.Refresh()
}

// executeSelected выполняет выбранную команду
func (m *uiModel) executeSelected() {
	if m.currentItem >= 0 && m.currentItem < len(m.filtered) {
		selectedKey := m.filtered[m.currentItem].Title
		if cmd, exists := m.getSelectedApp(selectedKey); exists {
			m.executeCommand(cmd.Command)
			m.window.Close()
		}
	}
}

func (m *uiModel) getSelectedApp(
	selectedKey string,
) (
	model.App,
	bool,
) {
	// TODO крайне сомнительный способ получать приложение по тайтлу
	exists := false

	for _, app := range m.items {
		if app.Title == selectedKey {
			exists = true
			return app, exists
		}
	}

	return model.App{}, exists
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
		item.title.Color = color.White
		item.description.Color = color.White
	} else {
		item.background.FillColor = color.White
		item.title.Color = color.Black
		item.description.Color = color.Gray{0x80}
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
				item.title.Text = key.Title
				item.description.Text = key.Description

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

	// Обработка выбора из списка - только при клике мышью
	list.OnSelected = func(id widget.ListItemID) {
		if !m.ignoreSelection {
			m.currentItem = id
			m.executeSelected()
		}
	}

	return list
}

// CustomEntry - кастомное поле ввода, которое передает стрелки наверх
type CustomEntry struct {
	widget.Entry
	onArrowUp   func()
	onArrowDown func()
	onEnter     func()
}

// NewCustomEntry создает новое кастомное поле ввода
func NewCustomEntry() *CustomEntry {
	entry := &CustomEntry{}
	entry.ExtendBaseWidget(entry)
	// entry.Wrapping = fyne.TextTruncation
	return entry
}

// TypedKey обрабатывает нажатия клавиш
func (e *CustomEntry) TypedKey(key *fyne.KeyEvent) {
	switch key.Name {
	case fyne.KeyUp:
		if e.onArrowUp != nil {
			e.onArrowUp()
		}
		return // Предотвращаем стандартную обработку
	case fyne.KeyDown:
		if e.onArrowDown != nil {
			e.onArrowDown()
		}
		return // Предотвращаем стандартную обработку
	case fyne.KeyReturn, fyne.KeyEnter:
		if e.onEnter != nil {
			e.onEnter()
		} else {
			e.Entry.TypedKey(key)
		}
		return
	default:
		// Для остальных клавиш вызываем стандартный обработчик
		e.Entry.TypedKey(key)
	}
}

// SetOnArrowUp устанавливает обработчик стрелки вверх
func (e *CustomEntry) SetOnArrowUp(handler func()) {
	e.onArrowUp = handler
}

// SetOnArrowDown устанавливает обработчик стрелки вниз
func (e *CustomEntry) SetOnArrowDown(handler func()) {
	e.onArrowDown = handler
}

// SetOnEnter устанавливает обработчик Enter
func (e *CustomEntry) SetOnEnter(handler func()) {
	e.onEnter = handler
}

func StartUI(apps []model.App) {
	myApp := app.New()
	myWindow := myApp.NewWindow("Fast Launcher")
	myWindow.SetFixedSize(true)
	myWindow.Resize(fyne.NewSize(600, 400))
	myWindow.CenterOnScreen()

	// Создаём модель как в TUI версии
	m := &uiModel{
		items:  apps,
		window: myWindow,
	}

	// Используем кастомное поле ввода
	input := NewCustomEntry()
	input.SetPlaceHolder("Введите команду для поиска...")
	m.input = &input.Entry

	// Устанавливаем обработчики стрелок
	input.SetOnArrowUp(func() {
		m.moveSelection(-1)
	})
	input.SetOnArrowDown(func() {
		m.moveSelection(1)
	})
	input.SetOnEnter(func() {
		m.executeSelected()
	})

	// Создаем кастомный список
	list := m.createCustomList()
	m.list = list

	// Обработка ввода - используем fuzzy search из TUI версии
	input.OnChanged = func(text string) {
		m.updateList()
	}

	// Дополнительная обработка клавиш на уровне окна для Escape
	myWindow.Canvas().SetOnTypedKey(func(key *fyne.KeyEvent) {
		if key.Name == fyne.KeyEscape {
			myWindow.Close()
		}
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
