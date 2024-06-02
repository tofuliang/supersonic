package browsing

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/dweymouth/supersonic/backend"
	"github.com/dweymouth/supersonic/backend/mediaprovider"
	"github.com/dweymouth/supersonic/ui/controller"
	"github.com/dweymouth/supersonic/ui/layouts"
	myTheme "github.com/dweymouth/supersonic/ui/theme"
)

type Page interface {
	fyne.CanvasObject

	Save() SavedPage
	Reload()
	Route() controller.Route
}

type SavedPage interface {
	Restore() Page
}

// Searchable pages should implement this interface so their search bar can be focused by keyboard shortcut.
type Searchable interface {
	SearchWidget() fyne.Focusable
}

// Pages with selection should implement this interface to receive Ctrl+A events
type CanSelectAll interface {
	SelectAll()
}

// Pages that have one main scrollable view should implement this interface
// to receive callbacks from window-level keyboard scrolling (up/down)
type Scrollable interface {
	Scroll(amount float32)
}

type CanShowNowPlaying interface {
	OnSongChange(playing mediaprovider.MediaItem, lastScrobbledIfAny *mediaprovider.Track)
}

type CanShowPlayTime interface {
	OnPlayTimeUpdate(curTime, totalTime float64, seeked bool)
}

type CanShowPlayQueue interface {
	OnPlayQueueChange()
}

type BrowsingPane struct {
	widget.BaseWidget

	app *backend.App

	curPage Page

	home       *widget.Button
	forward    *widget.Button
	back       *widget.Button
	reload     *widget.Button
	history    []SavedPage
	historyIdx int

	settingsBtn      *widget.Button
	settingsMenu     *fyne.Menu
	navBtnsContainer *fyne.Container
	pageContainer    *fyne.Container
	container        *fyne.Container
	navBtnsPageMap   map[controller.PageName]fyne.Resource
}

func NewBrowsingPane(app *backend.App, contr *controller.Controller, onGoHome func()) *BrowsingPane {
	b := &BrowsingPane{app: app}
	b.ExtendBaseWidget(b)
	b.home = widget.NewButtonWithIcon("", theme.HomeIcon(), onGoHome)
	b.back = widget.NewButtonWithIcon("", theme.NavigateBackIcon(), b.GoBack)
	b.forward = widget.NewButtonWithIcon("", theme.NavigateNextIcon(), b.GoForward)
	b.reload = widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), b.Reload)
	b.app.PlaybackManager.OnSongChange(b.onSongChange)
	b.app.PlaybackManager.OnPlayTimeUpdate(b.onPlayTimeUpdate)
	b.app.PlaybackManager.OnQueueChange(b.onQueueChange)
	bkgrnd := myTheme.NewThemedRectangle(myTheme.ColorNamePageBackground)
	b.pageContainer = container.NewStack(bkgrnd, layout.NewSpacer())
	b.settingsBtn = widget.NewButtonWithIcon("", theme.SettingsIcon(), func() {
		p := widget.NewPopUpMenu(b.settingsMenu,
			fyne.CurrentApp().Driver().CanvasForObject(b.settingsBtn))
		p.ShowAtPosition(fyne.NewPos(b.Size().Width-p.MinSize().Width+4,
			b.navBtnsContainer.MinSize().Height+theme.Padding()))
	})
	quickSearchBtn := widget.NewButtonWithIcon("", theme.SearchIcon(), contr.ShowQuickSearch)
	b.settingsMenu = fyne.NewMenu("")
	b.navBtnsContainer = container.NewHBox()
	b.navBtnsPageMap = map[controller.PageName]fyne.Resource{}
	b.container = container.NewBorder(container.New(
		&layout.CustomPaddedLayout{LeftPadding: -5, RightPadding: -5},
		container.New(layouts.NewLeftMiddleRightLayout(0),
			container.NewHBox(b.home, b.back, b.forward, b.reload), b.navBtnsContainer,
			container.NewHBox(layout.NewSpacer(), quickSearchBtn, b.settingsBtn))),
		nil, nil, nil, b.pageContainer)
	b.updateHistoryButtons()
	return b
}

func (b *BrowsingPane) SetPage(p Page) {
	if p == nil {
		// special case to set a "blank page"
		// only used on logout, in conjunction with clearing the history
		b.pageContainer.Objects[1] = layout.NewSpacer()
		b.curPage = nil
		b.pageContainer.Refresh()
		b.updateNavBtnsColor(p)
	} else {
		oldPage := b.curPage
		if b.doSetPage(p) && oldPage != nil {
			b.addPageToHistory(oldPage, true)
			b.updateHistoryButtons()
		}
	}
}

func (b *BrowsingPane) ClearHistory() {
	b.history = nil
	b.historyIdx = 0
	b.updateHistoryButtons()
}

func (b *BrowsingPane) AddSettingsMenuItem(label string, action func()) {
	b.settingsMenu.Items = append(b.settingsMenu.Items,
		fyne.NewMenuItem(label, action))
}

func (b *BrowsingPane) AddSettingsMenuSeparator() {
	b.settingsMenu.Items = append(b.settingsMenu.Items,
		fyne.NewMenuItemSeparator())
}

func (b *BrowsingPane) AddNavigationButton(icon fyne.Resource, pageName controller.PageName, action func()) *widget.Button {
	// make a copy of the icon, because it can change the color
	browsingPaneIcon := theme.NewThemedResource(icon)
	btn := widget.NewButtonWithIcon("", browsingPaneIcon, action)
	b.navBtnsContainer.Add(btn)
	b.navBtnsPageMap[pageName] = browsingPaneIcon
	return btn
}

func (b *BrowsingPane) DisableNavigationButtons() {
	for _, obj := range b.navBtnsContainer.Objects {
		obj.(*widget.Button).Disable()
	}
}

func (b *BrowsingPane) EnableNavigationButtons() {
	for _, obj := range b.navBtnsContainer.Objects {
		obj.(*widget.Button).Enable()
	}
}

func (b *BrowsingPane) ActivateNavigationButton(num int) {
	if num < len(b.navBtnsContainer.Objects) {
		btn := b.navBtnsContainer.Objects[num].(*widget.Button)
		if !btn.Disabled() && !btn.Hidden {
			btn.OnTapped()
		}
	}
}

func (b *BrowsingPane) GetSearchBarIfAny() fyne.Focusable {
	if s, ok := b.curPage.(Searchable); ok {
		return s.SearchWidget()
	}
	return nil
}

func (b *BrowsingPane) SelectAll() {
	if s, ok := b.curPage.(CanSelectAll); ok {
		s.SelectAll()
	}
}

func (b *BrowsingPane) ScrollUp() {
	if s, ok := b.curPage.(Scrollable); ok {
		s.Scroll(-75)
	}
}

func (b *BrowsingPane) ScrollDown() {
	if s, ok := b.curPage.(Scrollable); ok {
		s.Scroll(75)
	}
}

func (b *BrowsingPane) doSetPage(p Page) bool {
	if b.curPage != nil && b.curPage.Route() == p.Route() {
		return false
	}
	// TODO: reset focus only if something inside the previous page had focus
	if c := fyne.CurrentApp().Driver().CanvasForObject(b); c != nil {
		c.Focus(nil)
	}
	b.curPage = p
	if np, ok := p.(CanShowNowPlaying); ok {
		// inform page of currently playing track
		np.OnSongChange(b.app.PlaybackManager.NowPlaying(), nil)
	}
	b.pageContainer.Remove(b.curPage)
	b.pageContainer.Objects[1] = p
	b.updateNavBtnsColor(p)
	b.Refresh()
	return true
}

func (b *BrowsingPane) onSongChange(song mediaprovider.MediaItem, lastScrobbledIfAny *mediaprovider.Track) {
	if b.curPage == nil {
		return
	}
	if p, ok := b.curPage.(CanShowNowPlaying); ok {
		p.OnSongChange(song, lastScrobbledIfAny)
	}
}

func (b *BrowsingPane) onPlayTimeUpdate(cur, total float64, seeked bool) {
	if b.curPage == nil {
		return
	}
	if p, ok := b.curPage.(CanShowPlayTime); ok {
		p.OnPlayTimeUpdate(cur, total, seeked)
	}
}

func (b *BrowsingPane) onQueueChange() {
	if b.curPage == nil {
		return
	}
	if p, ok := b.curPage.(CanShowPlayQueue); ok {
		p.OnPlayQueueChange()
	}
}

func (b *BrowsingPane) addPageToHistory(p Page, truncate bool) {
	if truncate {
		// allow garbage collection of pages that will be removed from the history
		for i := b.historyIdx; i < len(b.history); i++ {
			b.history[i] = nil
		}
		b.history = b.history[:b.historyIdx]
	}
	if b.historyIdx < len(b.history) {
		b.history[b.historyIdx] = p.Save()
	} else {
		b.history = append(b.history, p.Save())
	}
	b.historyIdx++
}

func (b *BrowsingPane) updateHistoryButtons() {
	if b.historyIdx > 0 {
		b.back.Enable()
	} else {
		b.back.Disable()
	}
	if b.historyIdx < len(b.history)-1 {
		b.forward.Enable()
	} else {
		b.forward.Disable()
	}
}

func (b *BrowsingPane) GoBack() {
	if b.historyIdx > 0 {
		// due to widget reuse between pages,
		// we must create the new page before calling addPageToHistory
		p := b.history[b.historyIdx-1].Restore()
		b.addPageToHistory(b.curPage, false)
		b.historyIdx -= 2
		b.doSetPage(p)
		b.updateHistoryButtons()
	}
}

func (b *BrowsingPane) GoForward() {
	if b.historyIdx < len(b.history)-1 {
		p := b.history[b.historyIdx+1].Restore()
		b.addPageToHistory(b.curPage, false)
		b.doSetPage(p)
		b.updateHistoryButtons()
	}
}

func (b *BrowsingPane) Reload() {
	if b.curPage != nil {
		b.curPage.Reload()
	}
}

func (b *BrowsingPane) CurrentPage() controller.Route {
	if b.curPage == nil {
		return controller.Route{Page: controller.None}
	}
	return b.curPage.Route()
}

func (b *BrowsingPane) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(b.container)
}

func (b *BrowsingPane) updateNavBtnsColor(p Page) {
	for pageName, icon := range b.navBtnsPageMap {
		if p != nil && pageName == p.Route().Page {
			icon.(*theme.ThemedResource).ColorName = theme.ColorNamePrimary
		} else {
			icon.(*theme.ThemedResource).ColorName = theme.ColorNameForeground
		}
	}
}
