package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/probeldev/fastlauncher/model"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// testUI drives the UI on a simulation screen. All reads of primitive state go
// through update() because tview primitives are not safe for concurrent use.
// Closures passed to update() only copy values: assertions run on the test
// goroutine, so a failed check is a clean FAIL instead of runtime.Goexit()
// killing the tview event loop goroutine and hanging the test.
type testUI struct {
	app    *tview.Application
	screen tcell.SimulationScreen
	m      *uiModel
	done   chan struct{}
}

func newTestUI(t *testing.T, apps []model.App, opts ...func(*tview.Application)) *testUI {
	t.Helper()

	setupTheme()
	m := &uiModel{stack: []frame{{items: apps}}}
	app := tview.NewApplication()
	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatalf("init simulation screen: %v", err)
	}
	screen.SetSize(100, 30)
	app.SetScreen(screen)
	for _, opt := range opts {
		opt(app)
	}

	root := m.buildUI(app)
	m.itemsPerPage = 10
	m.updateList()
	app.SetRoot(root, true)

	ui := &testUI{app: app, screen: screen, m: m, done: make(chan struct{})}
	go func() {
		_ = app.Run()
		close(ui.done)
	}()

	// Wait until the event loop is up and has processed at least one update.
	ui.update(t, func() {})

	return ui
}

// update runs f in the event loop and waits for it to finish.
func (u *testUI) update(t *testing.T, f func()) {
	t.Helper()

	finished := make(chan struct{})
	go func() {
		u.app.QueueUpdate(f)
		close(finished)
	}()

	select {
	case <-finished:
	case <-time.After(2 * time.Second):
		t.Fatal("event loop did not process the update in time")
	}
}

// waitFor polls cond (which must use update to read state) until it is true.
func (u *testUI) waitFor(t *testing.T, cond func() bool) {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		var ok bool
		u.update(t, func() { ok = cond() })
		if ok {
			return
		}
		time.Sleep(time.Millisecond)
	}

	t.Fatal("condition was not met in time")
}

// press queues a key event and waits until it has been handled.
func (u *testUI) press(t *testing.T, key tcell.Key, r rune) {
	t.Helper()

	u.app.QueueEvent(tcell.NewEventKey(key, r, tcell.ModNone))
	u.update(t, func() {})
}

func (u *testUI) stop() {
	u.app.Stop()
	<-u.done
}

// clickRow injects a full left click (down and up) on the given visible row of
// the list.
func (u *testUI) clickRow(t *testing.T, row int) {
	t.Helper()

	var x, y int
	u.update(t, func() {
		x, y, _, _ = u.m.list.GetInnerRect()
	})

	u.screen.InjectMouse(x, y+row, tcell.Button1, tcell.ModNone)
	u.screen.InjectMouse(x, y+row, tcell.ButtonNone, tcell.ModNone)
}

// state returns a snapshot of the uiModel fields used by assertions. When the
// application has already stopped, the event loop is gone, so the state is read
// directly: Run() has returned and there is no concurrent access.
func (u *testUI) state(t *testing.T) uiState {
	t.Helper()

	snapshot := func() uiState {
		return uiState{
			stackLen:   len(u.m.stack),
			listCount:  u.m.list.GetItemCount(),
			current:    u.m.list.GetCurrentItem(),
			page:       u.m.currentFrame().page,
			inputText:  u.m.input.GetText(),
			focusInput: u.app.GetFocus() == u.m.input,
			pagesText:  u.m.pages.GetText(true),
		}
	}

	if u.isStopped() {
		return snapshot()
	}

	var s uiState
	u.update(t, func() { s = snapshot() })

	return s
}

type uiState struct {
	stackLen   int
	listCount  int
	current    int
	page       int
	inputText  string
	focusInput bool
	pagesText  string
}

// stoppedWithin reports whether the application stopped within two seconds.
func (u *testUI) stoppedWithin() bool {
	select {
	case <-u.done:
		return true
	case <-time.After(2 * time.Second):
		return false
	}
}

// isStopped reports the current state without waiting.
func (u *testUI) isStopped() bool {
	select {
	case <-u.done:
		return true
	default:
		return false
	}
}

// waitForFile waits until a command has created path.
func waitForFile(t *testing.T, path string) {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(path); err == nil {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}

	t.Fatalf("expected the command to create %s", path)
}

func TestEnterOnGroupOpensSubmenu(t *testing.T) {
	apps := []model.App{
		{Title: "Dev", Command: "ignored", Items: []model.App{{Title: "Project", Command: "true"}}},
	}
	ui := newTestUI(t, apps)
	defer ui.stop()

	ui.press(t, tcell.KeyEnter, 0)
	ui.waitFor(t, func() bool { return len(ui.m.stack) == 2 })

	s := ui.state(t)
	if !s.focusInput {
		t.Error("focus must stay in the input field after entering a submenu")
	}
	if s.inputText != "" {
		t.Errorf("search must be cleared on entering a submenu, got %q", s.inputText)
	}
	if s.listCount != 1 {
		t.Errorf("expected 1 nested item, got %d", s.listCount)
	}
}

func TestEnterOnDeadItemKeepsSearch(t *testing.T) {
	apps := []model.App{
		{Title: "Dead", Description: "no command and no items"},
		{Title: "Alive", Command: "true"},
	}
	ui := newTestUI(t, apps)
	defer ui.stop()

	ui.press(t, tcell.KeyEnter, 0)
	ui.press(t, tcell.KeyRune, 'A')

	ui.waitFor(t, func() bool { return ui.m.input.GetText() == "A" })

	s := ui.state(t)
	if !s.focusInput {
		t.Error("focus must stay in the input field after Enter on a dead item")
	}
	if s.stackLen != 1 {
		t.Errorf("stack must not change, got depth %d", s.stackLen)
	}
}

func TestEscGoesBackToRoot(t *testing.T) {
	apps := []model.App{
		{
			Title:   "Dev",
			Command: "ignored",
			Items: []model.App{{
				Title:   "Nested",
				Command: "ignored",
				Items:   []model.App{{Title: "Leaf", Command: "true"}},
			}},
		},
	}
	ui := newTestUI(t, apps)
	defer ui.stop()

	ui.press(t, tcell.KeyEnter, 0)
	ui.waitFor(t, func() bool { return len(ui.m.stack) == 2 })

	ui.press(t, tcell.KeyEnter, 0)
	ui.waitFor(t, func() bool { return len(ui.m.stack) == 3 })

	ui.press(t, tcell.KeyEscape, 0)
	ui.waitFor(t, func() bool { return len(ui.m.stack) == 2 })

	ui.press(t, tcell.KeyEscape, 0)
	ui.waitFor(t, func() bool { return len(ui.m.stack) == 1 })

	if s := ui.state(t); !s.focusInput {
		t.Error("focus must return to the input field after Esc")
	}
}

// A click must activate the clicked row, not the previously selected one:
// tview calls the list's selected handler before updating its current item.
func TestMouseClickActivatesClickedItem(t *testing.T) {
	withMouse := func(app *tview.Application) { app.EnableMouse(true) }

	t.Run("click on leaf below a group", func(t *testing.T) {
		// The marker proves that the clicked leaf's command was actually run,
		// not only that the application was stopped.
		marker := filepath.Join(t.TempDir(), "clicked")

		apps := []model.App{
			{Title: "Group", Command: "ignored", Items: []model.App{{Title: "Child", Command: "true"}}},
			{Title: "Leaf", Command: fmt.Sprintf("touch %q", marker)},
		}
		ui := newTestUI(t, apps, withMouse)
		defer ui.stop()

		if s := ui.state(t); s.current != 0 {
			t.Fatalf("expected the first item to be selected, got %d", s.current)
		}

		ui.clickRow(t, 1)

		if !ui.stoppedWithin() {
			t.Fatal("click on the leaf did not activate it")
		}
		if s := ui.state(t); s.stackLen != 1 {
			t.Fatalf("the click activated the group instead of the leaf, stack depth %d", s.stackLen)
		}
		waitForFile(t, marker)
	})

	t.Run("click on group above a leaf", func(t *testing.T) {
		apps := []model.App{
			{Title: "Group", Command: "ignored", Items: []model.App{{Title: "Child", Command: "true"}}},
			{Title: "Leaf", Command: "true"},
		}
		ui := newTestUI(t, apps, withMouse)
		defer ui.stop()

		ui.update(t, func() { ui.m.list.SetCurrentItem(1) })

		ui.clickRow(t, 0)

		ui.waitFor(t, func() bool { return len(ui.m.stack) == 2 })

		if ui.isStopped() {
			t.Fatal("the click activated the leaf instead of the group")
		}
	})

	// Esc must return to the clicked group even when another row was selected
	// before the click: the saved index is the clicked one, not the previous one.
	t.Run("click on group below the selection", func(t *testing.T) {
		apps := []model.App{
			{Title: "One", Command: "true"},
			{Title: "Two", Command: "true"},
			{Title: "Group", Command: "ignored", Items: []model.App{{Title: "Child", Command: "true"}}},
			{Title: "Four", Command: "true"},
		}
		ui := newTestUI(t, apps, withMouse)
		defer ui.stop()

		ui.update(t, func() { ui.m.list.SetCurrentItem(1) })

		ui.clickRow(t, 2)
		ui.waitFor(t, func() bool { return len(ui.m.stack) == 2 })

		ui.press(t, tcell.KeyEscape, 0)
		ui.waitFor(t, func() bool { return len(ui.m.stack) == 1 })

		if s := ui.state(t); s.current != 2 {
			t.Errorf("Esc must restore the clicked item, got %d", s.current)
		}
	})

	// On the second page the clicked row index must be offset by the page.
	t.Run("click on the second page", func(t *testing.T) {
		apps := make([]model.App, 0, 40)
		apps = append(apps, model.App{
			Title:   "Group",
			Command: "ignored",
			Items:   []model.App{{Title: "Child", Command: "true"}},
		})
		for i := 1; i < 40; i++ {
			apps = append(apps, model.App{Title: "Leaf", Command: "true"})
		}

		ui := newTestUI(t, apps, withMouse)
		defer ui.stop()

		ui.press(t, tcell.KeyRight, 0)

		s := ui.state(t)
		if s.page != 1 {
			t.Fatalf("expected page 2, got %d", s.page+1)
		}
		if s.listCount == 0 || s.listCount >= len(apps) {
			t.Fatalf("expected a partial second page, got %d rows", s.listCount)
		}

		ui.clickRow(t, 0)

		if !ui.stoppedWithin() {
			t.Fatal("click on the second page did not activate the leaf there")
		}
		if s := ui.state(t); s.stackLen != 1 {
			t.Fatalf("the click activated an item from the first page, stack depth %d", s.stackLen)
		}
	})
}

func TestEmptyListShowsNoItems(t *testing.T) {
	ui := newTestUI(t, nil)
	defer ui.stop()

	s := ui.state(t)
	if !strings.Contains(s.pagesText, "No items") {
		t.Errorf("expected %q in the status line, got %q", "No items", s.pagesText)
	}
	if strings.Contains(s.pagesText, "/0") {
		t.Errorf("status line must not show zero pages, got %q", s.pagesText)
	}
	if s.listCount != 0 {
		t.Errorf("expected an empty list, got %d items", s.listCount)
	}
}

func TestEscRestoresSelectedItem(t *testing.T) {
	apps := []model.App{
		{Title: "One", Command: "true"},
		{Title: "Two", Command: "true"},
		{Title: "Group", Command: "ignored", Items: []model.App{{Title: "Child", Command: "true"}}},
		{Title: "Four", Command: "true"},
	}
	ui := newTestUI(t, apps)
	defer ui.stop()

	ui.press(t, tcell.KeyDown, 0)
	ui.press(t, tcell.KeyDown, 0)

	if s := ui.state(t); s.current != 2 {
		t.Fatalf("expected the group to be selected, got %d", s.current)
	}

	ui.press(t, tcell.KeyEnter, 0)
	ui.waitFor(t, func() bool { return len(ui.m.stack) == 2 })

	ui.press(t, tcell.KeyEscape, 0)
	ui.waitFor(t, func() bool { return len(ui.m.stack) == 1 })

	if s := ui.state(t); s.current != 2 {
		t.Errorf("selected item was not restored, got %d", s.current)
	}
}
