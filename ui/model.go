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

	// Устанавливаем текущий элемент на первый в списке, если список не пуст
	if m.list.GetItemCount() > 0 && m.list.GetCurrentItem() < 0 {
		m.list.SetCurrentItem(0)
	}

	// Обновляем индикатор пагинации
	pageText := fmt.Sprintf("Страница %d/%d (←/→)", m.currentPage+1, totalPages)
	m.pages.SetText(pageText)
}

// updateItemsPerPage вычисляет количество элементов на странице
func (m *uiModel) updateItemsPerPage() {
	_, _, _, height := m.list.GetRect()
	// Учитываем высоту поля поиска (1) и пагинатора (1)
	m.itemsPerPage = height - 2
	if m.itemsPerPage < 1 {
		m.itemsPerPage = 1
	}
}

func StartUi(apps []model.App) {
	// Создаём модель
	m := &uiModel{
		items:       make([]item, len(apps)),
		currentPage: 0,
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
			// Отладка для проверки вызова
			fmt.Println("SetSelectedFunc called, Index:", index, "MainText:", mainText)
			// Запускаем команду
			actualIndex := m.currentPage*m.itemsPerPage + index
			filtered := m.filterItems(m.input.GetText())
			if actualIndex < len(filtered) {
				fmt.Println("Actual index:", actualIndex, "Filtered items:", len(filtered), "Command:", filtered[actualIndex].command)
				runner, err := apprunner.GetAppRunner(apprunner.OsLinux)
				if err != nil {
					fmt.Println("GetAppRunner error:", err)
					log.Println("GetAppRunner error:", err)
					return
				}
				err = runner.Run(filtered[actualIndex].command)
				if err != nil {
					fmt.Println("Run error:", err)
					log.Println("Run error:", err)
					return
				}
			} else {
				fmt.Println("Invalid index:", actualIndex, "Filtered length:", len(filtered))
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
		m.updateItemsPerPage()
		m.updateList()
		return x, y, width, height
	})

	// Настраиваем клавиши
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// fmt.Println("Key pressed:", event.Key(), "Focus:", app.GetFocus())
		switch event.Key() {
		case tcell.KeyLeft:
			m.currentPage--
			m.updateList()
			return nil
		case tcell.KeyRight:
			m.currentPage++
			m.updateList()
			return nil
		case tcell.KeyUp:
			// Переключаем фокус на список, если он не в фокусе
			if app.GetFocus() != m.list {
				app.SetFocus(m.list)
			}
			// Перемещаем курсор вверх
			current := m.list.GetCurrentItem()
			if current > 0 {
				m.list.SetCurrentItem(current - 1)
			}
			return nil
		case tcell.KeyDown:
			// Переключаем фокус на список, если он не в фокусе
			if app.GetFocus() != m.list {
				app.SetFocus(m.list)
			}
			// Перемещаем курсор вниз
			current := m.list.GetCurrentItem()
			if current < m.list.GetItemCount()-1 {
				m.list.SetCurrentItem(current + 1)
			}
			return nil
		case tcell.KeyCtrlC, tcell.KeyEscape:
			fmt.Println("Exiting application")
			app.Stop()
			return nil
		case tcell.KeyEnter:
			fmt.Println("Enter pressed, Focus:", app.GetFocus(), "ItemCount:", m.list.GetItemCount(), "CurrentItem:", m.list.GetCurrentItem())
			if m.list.GetItemCount() == 0 {
				fmt.Println("No items in list")
				return nil
			}
			if app.GetFocus() == m.input {
				// Переключаем фокус на список
				app.SetFocus(m.list)
				// Вызываем SetSelectedFunc для текущего элемента
				current := m.list.GetCurrentItem()
				if current >= 0 && current < m.list.GetItemCount() {
					mainText, secondaryText := m.list.GetItemText(current)
					fmt.Println("Calling SetSelectedFunc from input focus")
					m.list.GetSelectedFunc()(current, mainText, secondaryText, 0)
				} else {
					fmt.Println("Invalid current item:", current)
				}
				return nil
			}
			// Позволяем списку обработать Enter
			fmt.Println("Forwarding Enter to list")
			return event
		}
		return event
	})

	// Инициализируем itemsPerPage и список
	m.updateItemsPerPage()
	m.updateList()

	// Запускаем приложение
	fmt.Println("Starting application")
	if err := app.SetRoot(flex, true).Run(); err != nil {
		fmt.Println("Error running program:", err)
		log.Println("Error running program:", err)
		os.Exit(1)
	}
}
