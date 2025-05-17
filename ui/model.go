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
	// Высота всего экрана минус фиксированные элементы (поле ввода, пагинация и padding)
	m.itemsPerPage = height - 4 // 1 строка для ввода, 1 строка для пагинации, 2 строки для padding сверху и снизу
	if m.itemsPerPage < 1 {
		m.itemsPerPage = 1
	}
}

func StartUi(apps []model.App) {
	// Настраиваем стили tview для использования цветов терминала
	tview.Styles = tview.Theme{
		PrimitiveBackgroundColor:    tcell.ColorDefault,
		ContrastBackgroundColor:     tcell.ColorDefault,
		MoreContrastBackgroundColor: tcell.ColorDefault,
		BorderColor:                 tcell.ColorDefault,
		TitleColor:                  tcell.ColorDefault,
		GraphicsColor:               tcell.ColorDefault,
		PrimaryTextColor:            tcell.ColorDefault,
		SecondaryTextColor:          tcell.ColorDefault,
		TertiaryTextColor:           tcell.ColorDefault,
		InverseTextColor:            tcell.ColorDefault,
		ContrastSecondaryTextColor:  tcell.ColorDefault,
	}

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
		SetMainTextStyle(tcell.StyleDefault.Foreground(tcell.ColorDefault).Background(tcell.ColorDefault)).
		SetSelectedStyle(tcell.StyleDefault.Foreground(tcell.ColorDefault).Background(tcell.ColorDefault).Reverse(true)).
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

	// Создаём поле ввода с рамкой
	m.input = tview.NewInputField()
	m.input.SetLabel("Поиск: ").
		SetLabelStyle(tcell.StyleDefault.Foreground(tcell.ColorDefault)).
		SetFieldStyle(tcell.StyleDefault.Foreground(tcell.ColorDefault).Background(tcell.ColorDefault)).
		SetBorder(true).
		SetBorderStyle(tcell.StyleDefault.Foreground(tcell.ColorDefault))
	m.input.SetChangedFunc(func(text string) {
		m.currentPage = 0 // Сбрасываем страницу при изменении поиска
		m.updateList()
	})

	// Создаём индикатор пагинации
	m.pages = tview.NewTextView().
		SetTextAlign(tview.AlignLeft).
		SetTextStyle(tcell.StyleDefault.Foreground(tcell.ColorDefault).Background(tcell.ColorDefault))

	// Компоновка с padding
	innerFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(m.input, 3, 1, true). // Увеличиваем высоту для рамки (1 строка текста + 2 строки рамки)
		AddItem(m.list, 0, 1, false).
		AddItem(m.pages, 1, 1, false)

	// Добавляем padding (1 строка сверху и снизу, 2 столбца слева и справа)
	outerFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(tview.NewBox(), 1, 0, false). // Padding сверху
		AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(tview.NewBox(), 2, 0, false). // Padding слева
			AddItem(innerFlex, 0, 1, true).
			AddItem(tview.NewBox(), 2, 0, false), // Padding справа
							0, 1, true).
		AddItem(tview.NewBox(), 1, 0, false) // Padding снизу

	// Настраиваем обработку изменения размера через SetDrawFunc
	outerFlex.SetDrawFunc(func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
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
			// Переключаем фокус на список, если он не в фокусе
			if app.GetFocus() != m.list {
				app.SetFocus(m.list)
				// Устанавливаем первый элемент, если список только получил фокус
				if m.list.GetItemCount() > 0 && m.list.GetCurrentItem() < 0 {
					m.list.SetCurrentItem(0)
				}
			}
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
	if err := app.SetRoot(outerFlex, true).Run(); err != nil {
		log.Println("Error running program:", err)
		os.Exit(1)
	}
}
