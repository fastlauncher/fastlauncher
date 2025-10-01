package ui

import (
	"log"
	"os/exec"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
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
	if m.list != nil {
		m.list.Refresh()
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

	// Список как в Fyne версии, но с данными из TUI
	list := widget.NewList(
		func() int {
			return len(m.filtered)
		},
		func() fyne.CanvasObject {
			return container.NewVBox(
				widget.NewLabel("template"),
				widget.NewLabel("description"),
			)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			if i < len(m.filtered) {
				key := m.filtered[i]
				labels := o.(*fyne.Container).Objects
				labels[0].(*widget.Label).SetText(key)
				labels[1].(*widget.Label).SetText(m.commands[key])
			}
		},
	)
	m.list = list

	// Обработка выбора из списка (комбинация обеих версий)
	list.OnSelected = func(id widget.ListItemID) {
		if id < len(m.filtered) {
			selectedKey := m.filtered[id]
			if cmd, exists := m.commands[selectedKey]; exists {
				m.executeCommand(cmd)
				myWindow.Close()
			}
		}
	}

	// Обработка ввода - используем fuzzy search из TUI версии
	input.OnChanged = func(text string) {
		m.updateList()
	}

	// Обработка Enter - комбинация обеих версий
	input.OnSubmitted = func(cmd string) {
		// Сначала пытаемся найти точное совпадение
		if command, exists := m.commands[cmd]; exists {
			m.executeCommand(command)
			myWindow.Close()
			return
		}

		// Если есть отфильтрованные элементы, берем первый
		if len(m.filtered) > 0 {
			selectedKey := m.filtered[0]
			if command, exists := m.commands[selectedKey]; exists {
				m.executeCommand(command)
				myWindow.Close()
				return
			}
		}

		// Если ничего не найдено, выполняем как есть (из Fyne версии)
		m.executeCommand(cmd)
		myWindow.Close()
	}

	// Компоновка как в Fyne версии
	content := container.NewBorder(
		input, // сверху - поле ввода
		nil, nil, nil,
		list, // по центру - список подсказок
	)

	myWindow.SetContent(content)

	// Фокус на поле ввода при открытии
	myWindow.Canvas().Focus(input)

	// Обработка Escape для закрытия (из Fyne версии)
	myWindow.Canvas().SetOnTypedKey(func(e *fyne.KeyEvent) {
		if e.Name == fyne.KeyEscape {
			myWindow.Close()
		}
	})

	// Инициализируем список
	m.updateList()

	myWindow.ShowAndRun()
}
