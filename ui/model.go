// Package ui...
package ui

import (
	"errors"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"

	filteritems "github.com/probeldev/fastlauncher/filterItems"
	"github.com/probeldev/fastlauncher/model"
	"github.com/probeldev/fastlauncher/pkg/apprunner"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const breadcrumbsSeparator = " › "

// frame is one level of the nested menu navigation.
type frame struct {
	title    string
	items    []model.App
	page     int
	query    string
	selected int
}

type uiModel struct {
	stack        []frame
	list         *tview.List
	input        *tview.InputField
	pages        *tview.TextView
	itemsPerPage int
	lastWidth    int
	lastHeight   int
}

func (m *uiModel) currentFrame() *frame {
	return &m.stack[len(m.stack)-1]
}

// filteredItems returns the items of the current frame matching its query.
func (m *uiModel) filteredItems() []model.App {
	current := m.currentFrame()

	return filteritems.FilterItems(current.items, current.query)
}

// updateList обновляет содержимое списка
func (m *uiModel) updateList() {
	current := m.list.GetCurrentItem() // Сохраняем текущий элемент
	m.list.Clear()
	filtered := m.filteredItems()
	totalItems := len(filtered)
	totalPages := (totalItems + m.itemsPerPage - 1) / m.itemsPerPage

	cur := m.currentFrame()
	if cur.page >= totalPages {
		cur.page = totalPages - 1
	}
	if cur.page < 0 {
		cur.page = 0
	}

	start := cur.page * m.itemsPerPage
	end := start + m.itemsPerPage
	end = min(end, totalItems)

	for i := start; i < end; i++ {
		m.list.AddItem(filtered[i].Title, "", 0, nil)
	}

	// Восстанавливаем текущий элемент, если он в пределах нового списка
	if current >= 0 && current < end-start {
		m.list.SetCurrentItem(current)
	} else if m.list.GetItemCount() > 0 {
		m.list.SetCurrentItem(0)
	}

	m.updatePagesText(totalPages)
}

// updatePagesText обновляет хлебные крошки и индикатор пагинации.
func (m *uiModel) updatePagesText(totalPages int) {
	pageText := "No items"
	if totalPages > 0 {
		pageText = fmt.Sprintf("Page %d/%d (←/→)", m.currentFrame().page+1, totalPages)
	}
	if title := m.currentFrame().title; title != "" {
		pageText = title + "   |   " + pageText
	}

	m.pages.SetText(pageText)
}

// updateItemsPerPage вычисляет количество элементов на странице
func (m *uiModel) updateItemsPerPage(height int) {
	// Высота всего экрана минус фиксированные элементы (поле ввода, пагинация и padding)
	m.itemsPerPage = height - 4 // 1 строка для ввода, 1 строка для пагинации, 2 строки для padding сверху и снизу
	m.itemsPerPage = max(m.itemsPerPage, 1)
}

// selectNextPage переходит на следующую страницу текущего кадра.
func (m *uiModel) selectNextPage() {
	filtered := m.filteredItems()
	totalPages := (len(filtered) + m.itemsPerPage - 1) / m.itemsPerPage
	if m.currentFrame().page < totalPages-1 {
		m.currentFrame().page++
		m.updateList()
	}
}

// selectPrevPage переходит на предыдущую страницу текущего кадра.
func (m *uiModel) selectPrevPage() {
	if m.currentFrame().page > 0 {
		m.currentFrame().page--
		m.updateList()
	}
}

// restoreSelection выделяет сохранённый пункт текущего кадра.
func (m *uiModel) restoreSelection(index int) {
	index = max(index, 0)
	index = min(index, m.list.GetItemCount()-1)
	if index >= 0 {
		m.list.SetCurrentItem(index)
	}
}

// moveSelection перемещает выделение внутри текущей страницы.
func (m *uiModel) moveSelection(delta int) {
	current := m.list.GetCurrentItem()
	next := current + delta
	if next < 0 || next >= m.list.GetItemCount() {
		return
	}
	m.list.SetCurrentItem(next)
}

// childTitle строит хлебные крошки для вложенного кадра.
func childTitle(parentTitle, childTitle string) string {
	if parentTitle == "" {
		return childTitle
	}

	return parentTitle + breadcrumbsSeparator + childTitle
}

// enterSubmenu открывает кадр вложенного меню. index — позиция пункта в списке,
// её запоминаем, чтобы вернуться к тому же пункту по Esc.
func (m *uiModel) enterSubmenu(app *tview.Application, item *model.App, index int) {
	parent := m.currentFrame()
	parent.selected = index
	title := childTitle(parent.title, item.Title)

	m.stack = append(m.stack, frame{
		title: title,
		items: item.Items,
	})

	// Новый кадр начинается с пустого поиска и первой страницы. SetText("")
	// вызывает changed (и обновление списка) только если текст непустой,
	// поэтому при уже пустом поле список обновляем явно.
	if m.input.GetText() == "" {
		m.updateList()
	} else {
		m.input.SetText("")
	}
	m.list.SetCurrentItem(0)
	app.SetFocus(m.input)
}

// goBack возвращает на предыдущий кадр. Возвращает false на корневом кадре.
func (m *uiModel) goBack(app *tview.Application) bool {
	if len(m.stack) <= 1 {
		return false
	}

	m.stack = m.stack[:len(m.stack)-1]
	parent := m.currentFrame()
	savedPage := parent.page

	// SetText может не изменить текст, поэтому список обновляем явно.
	m.input.SetText(parent.query)
	parent.page = savedPage
	m.updateList()
	m.restoreSelection(parent.selected)
	app.SetFocus(m.input)

	return true
}

// activateItemAt активирует пункт списка по его индексу в текущей странице:
// открывает вложенное меню либо запускает команду и закрывает лаунчер.
// Индекс передаётся явно, потому что tview вызывает обработчик клика мыши до
// того, как обновит текущий элемент списка.
func (m *uiModel) activateItemAt(app *tview.Application, index int) {
	if index < 0 || index >= m.list.GetItemCount() {
		return
	}

	filtered := m.filteredItems()
	actualIndex := m.currentFrame().page*m.itemsPerPage + index

	if actualIndex >= len(filtered) {
		return
	}

	item := &filtered[actualIndex]
	if len(item.Items) > 0 {
		m.enterSubmenu(app, item, index)
		return
	}

	// Пункт без команды (например, группа, все дети которой не загрузились)
	// не запускаем и окно не закрываем.
	if strings.TrimSpace(item.Command) == "" {
		return
	}

	m.runCommand(item.Command)
	app.Stop()
}

// runCommand запускает команду пункта.
func (m *uiModel) runCommand(command string) {
	osForRunner, err := m.getRunnerOs()
	if err != nil {
		log.Println("GetRunnerOs error:", err)
		return
	}

	runner, err := apprunner.GetAppRunner(osForRunner)
	if err != nil {
		log.Println("GetAppRunner error:", err)
		return
	}

	if err := runner.Run(command); err != nil {
		log.Println("Run error:", err)
	}
}

// handleKeyEnter обрабатывает Enter в поле ввода: активирует выделенный пункт,
// не переводя фокус на список. Фокус остаётся в поле, чтобы поиск продолжал
// работать, в том числе если пункт оказался пустым и ничего не запустил.
func (m *uiModel) handleKeyEnter(app *tview.Application) *tcell.EventKey {
	m.activateItemAt(app, m.list.GetCurrentItem())

	return nil
}

func StartUI(apps []model.App) {
	setupTheme()

	// Создаём модель
	m := &uiModel{
		stack: []frame{{items: apps}},
	}

	// Создаём приложение
	app := tview.NewApplication()
	root := m.buildUI(app)

	// Инициализируем itemsPerPage и список
	m.itemsPerPage = 10 // Начальное значение, будет обновлено при первом вызове SetDrawFunc
	m.updateList()

	// Запускаем приложение
	if err := app.SetRoot(root, true).Run(); err != nil {
		log.Println("Error running program:", err)
		os.Exit(1)
	}
}

// buildUI создаёт примитивы интерфейса и возвращает корневой элемент.
func (m *uiModel) buildUI(app *tview.Application) tview.Primitive {
	// Создаём список
	m.list = tview.NewList().
		ShowSecondaryText(false).
		SetMainTextStyle(tcell.StyleDefault.Foreground(tcell.ColorDefault).Background(tcell.ColorDefault)).
		SetSelectedStyle(tcell.StyleDefault.Foreground(tcell.ColorDefault).Background(tcell.ColorDefault).Reverse(true)).
		SetSelectedFunc(func(index int, _, _ string, _ rune) {
			m.activateItemAt(app, index)
		})

	// Создаём поле ввода с рамкой
	m.input = tview.NewInputField()
	m.input.SetLabel("Search: ").
		SetLabelStyle(tcell.StyleDefault.Foreground(tcell.ColorDefault)).
		SetFieldStyle(tcell.StyleDefault.Foreground(tcell.ColorDefault).Background(tcell.ColorDefault)).
		SetBorder(true).
		SetBorderStyle(tcell.StyleDefault.Foreground(tcell.ColorDefault))
	m.input.SetChangedFunc(func(text string) {
		frame := m.currentFrame()
		frame.query = text
		frame.page = 0 // Сбрасываем страницу при изменении поиска
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
			m.selectPrevPage()
			return nil
		case tcell.KeyRight:
			m.selectNextPage()
			return nil
		case tcell.KeyUp, tcell.KeyDown:
			delta := 1
			if event.Key() == tcell.KeyUp {
				delta = -1
			}
			m.moveSelection(delta)
			return nil
		case tcell.KeyCtrlC:
			app.Stop()
			return nil
		case tcell.KeyEscape:
			if !m.goBack(app) {
				app.Stop()
			}
			return nil
		case tcell.KeyEnter:
			if app.GetFocus() == m.input {
				return m.handleKeyEnter(app)
			}
			// Позволяем списку обработать Enter
			return event
		}
		return event
	})

	return outerFlex
}

// setupTheme настраивает стили tview для использования цветов терминала.
func setupTheme() {
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
}

func (m *uiModel) getRunnerOs() (string, error) {
	currentOs := runtime.GOOS
	switch currentOs {
	case "darwin":
		return apprunner.OsMacOs, nil
	case "linux":
		return apprunner.OsLinux, nil
	}

	return "", errors.New("OS is not support")
}
