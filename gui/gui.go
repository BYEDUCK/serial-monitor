package gui

import (
	"fmt"
	"reflect"

	"byeduck.com/serial-monitor/utils"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

const (
	Text string = "TEXT"
	Plot string = "PLOT"
)

const (
	MAX_MSG_DISPLAY_SIZE = 30

	PARAGRAPH_HEIGHT = 3
	LIST_ELEM_HEIGHT = 1
)

func GetAvailableModes() []string {
	return []string{Text, Plot}
}

type MainGui struct {
	BaudParagraph              *widgets.Paragraph
	DeviceParagraph            *widgets.Paragraph
	ReadTimeoutParagraph       *widgets.Paragraph
	LogsEnabledParagraph       *widgets.Paragraph
	TimestampsEnabledParagraph *widgets.Paragraph
	HexModeParagraph           *widgets.Paragraph
	WrittenDataParagraph       *widgets.Paragraph
	ReadDataParagraph          *widgets.Paragraph
	PauseParagraph             *widgets.Paragraph
	FollowModeParagraph        *widgets.Paragraph
	InboxList                  *widgets.List
	InboxPlot                  *widgets.Plot
	InputParagraph             *widgets.Paragraph
}

func NewMainGui(mode string, fullScreen bool) *MainGui {
	availableWidth, availableHeight := ui.TerminalDimensions()
	mainStartX := 0
	var mainEndX int
	if fullScreen {
		mainEndX = availableWidth
	} else {
		mainEndX = int(float32(availableWidth)*0.75) - 1
	}
	mainStartY := 0
	var mainEndY int
	if mode == Text {
		mainEndY = MAX_MSG_DISPLAY_SIZE*LIST_ELEM_HEIGHT + 2
		if mainEndY > availableHeight {
			mainEndY = availableHeight - 20
		}
	} else if mode == Plot {
		mainEndY = int(float32(availableHeight) * 0.5)
	}

	var followModeParagraph *widgets.Paragraph
	var baudParagraph *widgets.Paragraph
	var deviceParagraph *widgets.Paragraph
	var readTimeoutParagraph *widgets.Paragraph
	var logsEnabledParagraph *widgets.Paragraph
	var timestampsEnabledParagraph *widgets.Paragraph
	var hexModeParagraph *widgets.Paragraph
	var writtenDataParagraph *widgets.Paragraph
	var readDataParagraph *widgets.Paragraph
	var pauseParagraph *widgets.Paragraph

	configCount := 0
	const configHeight = 1
	if !fullScreen {
		configWidth := int(float32(availableWidth)*0.25) - 1
		configStartX := mainEndX + 1
		configEndX := configStartX + configWidth
		baudParagraph = widgets.NewParagraph()
		baudParagraph.SetRect(configStartX, (configCount * PARAGRAPH_HEIGHT), configEndX, ((configCount + configHeight) * PARAGRAPH_HEIGHT))
		configCount++
		deviceParagraph = widgets.NewParagraph()
		deviceParagraph.SetRect(configStartX, (configCount * PARAGRAPH_HEIGHT), configEndX, ((configCount + configHeight) * PARAGRAPH_HEIGHT))
		configCount++
		readTimeoutParagraph = widgets.NewParagraph()
		readTimeoutParagraph.SetRect(configStartX, (configCount * PARAGRAPH_HEIGHT), configEndX, ((configCount + configHeight) * PARAGRAPH_HEIGHT))
		configCount++
		logsEnabledParagraph = widgets.NewParagraph()
		logsEnabledParagraph.SetRect(configStartX, (configCount * PARAGRAPH_HEIGHT), configEndX, ((configCount + configHeight) * PARAGRAPH_HEIGHT))
		configCount++
		writtenDataParagraph = widgets.NewParagraph()
		writtenDataParagraph.SetRect(configStartX, (configCount * PARAGRAPH_HEIGHT), configEndX, ((configCount + configHeight) * PARAGRAPH_HEIGHT))
		configCount++
		readDataParagraph = widgets.NewParagraph()
		readDataParagraph.SetRect(configStartX, (configCount * PARAGRAPH_HEIGHT), configEndX, ((configCount + configHeight) * PARAGRAPH_HEIGHT))
		configCount++
		pauseParagraph = widgets.NewParagraph()
		pauseParagraph.SetRect(configStartX, (configCount * PARAGRAPH_HEIGHT), configEndX, ((configCount + configHeight) * PARAGRAPH_HEIGHT))
		configCount++

		if mode == Text {
			hexModeParagraph = widgets.NewParagraph()
			hexModeParagraph.SetRect(configStartX, (configCount * PARAGRAPH_HEIGHT), configEndX, ((configCount + configHeight) * PARAGRAPH_HEIGHT))
			configCount++
			timestampsEnabledParagraph = widgets.NewParagraph()
			timestampsEnabledParagraph.SetRect(configStartX, (configCount * PARAGRAPH_HEIGHT), configEndX, ((configCount + configHeight) * PARAGRAPH_HEIGHT))
			configCount++
			followModeParagraph = widgets.NewParagraph()
			followModeParagraph.SetRect(configStartX, (configCount * PARAGRAPH_HEIGHT), configEndX, ((configCount + configHeight) * PARAGRAPH_HEIGHT))
		}
	}

	var inboxList *widgets.List
	var inboxPlot *widgets.Plot

	if mode == Text {
		inboxList = widgets.NewList()
		inboxList.TextStyle = ui.NewStyle(ui.ColorWhite)
		inboxList.WrapText = true
		inboxList.SetRect(mainStartX, mainStartY, mainEndX, mainEndY)
		inboxList.Title = fmt.Sprintf("%s(%d)", "IN messages", MAX_MSG_DISPLAY_SIZE)
	} else if mode == Plot {
		inboxPlot = widgets.NewPlot()
		inboxPlot.Title = "IN"
		inboxPlot.LineColors[0] = ui.ColorYellow
		inboxPlot.PlotType = widgets.LineChart
		inboxPlot.Marker = widgets.MarkerDot
		inboxPlot.SetRect(mainStartX, mainStartY, mainEndX, mainEndY)
	} else {
		panic("unknown gui mode")
	}

	inputWidget := widgets.NewParagraph()
	inputWidget.SetRect(0, (mainEndY + 1), mainEndX, (mainEndY + (2 * PARAGRAPH_HEIGHT) + 1))
	inputWidget.WrapText = true

	return &MainGui{
		BaudParagraph:              baudParagraph,
		DeviceParagraph:            deviceParagraph,
		ReadTimeoutParagraph:       readTimeoutParagraph,
		LogsEnabledParagraph:       logsEnabledParagraph,
		TimestampsEnabledParagraph: timestampsEnabledParagraph,
		HexModeParagraph:           hexModeParagraph,
		WrittenDataParagraph:       writtenDataParagraph,
		ReadDataParagraph:          readDataParagraph,
		PauseParagraph:             pauseParagraph,
		FollowModeParagraph:        followModeParagraph,
		InboxList:                  inboxList,
		InboxPlot:                  inboxPlot,
		InputParagraph:             inputWidget,
	}
}

func (g *MainGui) Render() {
	var guiWidgets []ui.Drawable
	appendWidgetIfNotNull := func(w ui.Drawable) {
		if !reflect.ValueOf(w).IsNil() {
			guiWidgets = append(guiWidgets, w)
		}
	}
	appendWidgetIfNotNull(g.BaudParagraph)
	appendWidgetIfNotNull(g.DeviceParagraph)
	appendWidgetIfNotNull(g.ReadTimeoutParagraph)
	appendWidgetIfNotNull(g.WrittenDataParagraph)
	appendWidgetIfNotNull(g.ReadDataParagraph)
	appendWidgetIfNotNull(g.InputParagraph)
	appendWidgetIfNotNull(g.PauseParagraph)
	appendWidgetIfNotNull(g.LogsEnabledParagraph)
	appendWidgetIfNotNull(g.TimestampsEnabledParagraph)
	appendWidgetIfNotNull(g.HexModeParagraph)
	appendWidgetIfNotNull(g.InboxList)
	appendWidgetIfNotNull(g.InboxPlot)
	appendWidgetIfNotNull(g.FollowModeParagraph)
	ui.Render(guiWidgets...)
}

func Init() {
	utils.Must("init ui", ui.Init())
}

func Close() {
	ui.Close()
}
