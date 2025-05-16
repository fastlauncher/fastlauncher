package ui

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/probeldev/fastlauncher/model"
	"github.com/probeldev/fastlauncher/pkg/apprunner"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type item struct {
	title   string
	command string
}

type uiModel struct {
	items        []item
	list         *tview.List
	input        *tview.InputField
	pages        *tview.TextView
	currentPage  int
	itemsPerPage int
	currentItem  int
	lastWidth    int
	lastHeight   int
}

// filterItems фильтрует элементы по запросу
func (m *uiModel) filterItems(query string) []item {
	if query == "" {
		return m.items
	}
	var filtered []item
	query = strings.ToLower(query)
	for _, item := range m.items {
		if strings.Contains(strings.ToLower(item.title), query) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

// updateList обновляет содержимое списка
func (m *uiModel) updateList() {
	current := m.list.GetCurrentItem() // Сохраняем текущий элемент
	m.list.Clear()
	filtered := m.filterItems(m.input.GetText())
	totalItems := len(filtered)
	totalPages := (totalItems + m.itemsPerPage - 1) / m.itemsPerPage
	if m.currentPage >= totalPages {
		m.currentPage = totalPages - 1
	}
	if m.currentPage < 0 {
		m.currentPage = 0
	}

	start := m.currentPage * m.itemsPerPage
	end := start + m.itemsPerPage
	if end > totalItems {
		end = totalItems
	}

	for i := start; i < end; i++ {
		m.list.AddItem(filtered[i].title, "", 0, nil)
	}

	// Восстанавливаем текущий элемент, если он в пределах нового списка
	if current >= 0 && current < end-start {
		m.list.SetCurrentItem(current)
	} else if m.list.GetItemCount() > 0 {
		m.list.SetCurrentItem(0)
	}

	// Обновляем индикатор пагинации
	pageText := fmt.Sprintf("Страница %d/%d (←/→)", m.currentPage+1, totalPages)
	m.pages.SetText(pageText)
}

// updateItemsPerPage вычисляет количество элементов на странице
func (m *uiModel) updateItemsPerPage(height int) {
	// Высота всего экрана минус фиксированные элементы (поле ввода и пагинация)
	m.itemsPerPage = height - 2 // 1 строка для ввода, 1 строка для пагинации
	if m.itemsPerPage < 1 {
		m.itemsPerPage = 1
	}
}

func StartUi(apps []model.App) {
	// Создаём модель
	m := &uiModel{
		items:       make([]item, len(apps)),
		currentPage: 0,
		currentItem: 0,
	}

	// Заполняем элементы
	for i, a := range apps {
		m.items[i] = item{
			title:   a.Title,
			command: a.Command,
		}
	}

	// Создаём приложение
	app := tview.NewApplication()

	// Создаём список
	m.list = tview.NewList().
		ShowSecondaryText(false).
		SetSelectedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
			// Запускаем команду
			actualIndex := m.currentPage*m.itemsPerPage + index
			filtered := m.filterItems(m.input.GetText())
			if actualIndex < len(filtered) {
				runner, err := apprunner.GetAppRunner(apprunner.OsLinux)
				if err != nil {
					log.Println("GetAppRunner error:", err)
					return
				}
				err = runner.Run(filtered[actualIndex].command)
				if err != nil {
					log.Println("Run error:", err)
					return
				}
			} else {
				log.Println("Invalid index:", actualIndex, "Filtered length:", len(filtered))
			}
			app.Stop()
		})

	// Создаём поле ввода
	m.input = tview.NewInputField().
		SetLabel("Поиск: ").
		SetFieldBackgroundColor(tview.Styles.PrimitiveBackgroundColor).
		SetChangedFunc(func(text string) {
			m.currentPage = 0 // Сбрасываем страницу при изменении поиска
			m.updateList()
		})

	// Создаём индикатор пагинации
	m.pages = tview.NewTextView().
		SetTextAlign(tview.AlignLeft).
		SetTextColor(tview.Styles.ContrastBackgroundColor)

	// Компоновка
	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(m.input, 1, 1, true). // Поле поиска в фокусе
		AddItem(m.list, 0, 1, false).
		AddItem(m.pages, 1, 1, false)

	// Настраиваем обработку изменения размера через SetDrawFunc
	flex.SetDrawFunc(func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
		// Проверяем, изменились ли размеры
		if m.lastWidth != width || m.lastHeight != height {
			m.lastWidth = width
			m.lastHeight = height
			m.updateItemsPerPage(height) // Используем полную высоту экрана
			m.updateList()
		}
		return x, y, width, height
	})

	// Настраиваем клавиши
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyLeft:
			if m.currentPage > 0 {
				m.currentPage--
				m.updateList()
			}
			return nil
		case tcell.KeyRight:
			filtered := m.filterItems(m.input.GetText())
			totalPages := (len(filtered) + m.itemsPerPage - 1) / m.itemsPerPage
			if m.currentPage < totalPages-1 {
				m.currentPage++
				m.updateList()
			}
			return nil
		case tcell.KeyUp, tcell.KeyDown:
			// Обрабатываем навигацию напрямую
			current := m.list.GetCurrentItem()
			if event.Key() == tcell.KeyUp {
				if current > 0 {
					m.list.SetCurrentItem(current - 1)
				}
			} else if event.Key() == tcell.KeyDown {
				if current < m.list.GetItemCount()-1 {
					m.list.SetCurrentItem(current + 1)
				}
			}
			return nil
		case tcell.KeyCtrlC, tcell.KeyEscape:
			app.Stop()
			return nil
		case tcell.KeyEnter:
			if m.list.GetItemCount() == 0 {
				return nil
			}
			if app.GetFocus() == m.input {
				// Переключаем фокус на список и выбираем первый элемент
				app.SetFocus(m.list)
				if m.list.GetItemCount() > 0 {
					current := m.list.GetCurrentItem()
					if current < 0 {
						current = 0
						m.list.SetCurrentItem(current)
					}
					mainText, secondaryText := m.list.GetItemText(current)
					m.list.GetSelectedFunc()(current, mainText, secondaryText, 0)
				}
				return nil
			}
			// Позволяем списку обработать Enter
			return event
		}
		return event
	})

	// Инициализируем itemsPerPage и список
	m.itemsPerPage = 10 // Начальное значение, будет обновлено при первом вызове SetDrawFunc
	m.updateList()

	// Запускаем приложение
	if err := app.SetRoot(flex, true).Run(); err != nil {
		log.Println("Error running program:", err)
		os.Exit(1)
	}
}
