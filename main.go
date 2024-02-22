package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"container/list"

	"byeduck.com/serial-monitor/gui"
	"byeduck.com/serial-monitor/utils"
	ui "github.com/gizak/termui/v3"
	"go.bug.st/serial"
)

const (
	INPUT_PREFIX                 = ">> "
	TEXT_NAVIGATION_INSTRUCTIONS = "i - enter input mode; h - hex mode; c - clear messages; s - print timestamps; j - scroll down; k - scroll up; t - scroll to top; b - scroll to bottom; f - enter/exit fallow mode, p - pause/unpause; m - change mode; z - zoom in/out; ESC - exit"
	PLOT_NAVIGATION_INSTRUCTIONS = "i - enter input mode; c - clear messages; p - pause/unpause; m - change mode; z - zoom in/out; ESC - exit"

	MAX_MSG_CAPACITY   = 10000
	MAX_POINT_CAPACITY = 200
	MSG_BUFF_SIZE      = 1000
)

var baud int
var readTimeoutMillieconds int
var logsEnabled bool
var guiMode string

var writtenBytes int64
var readBytes int64
var inputMode bool
var followMode bool
var paused bool
var fullScreen bool
var printTime bool
var hexMode bool

var serialPort serial.Port
var portName string
var messages *list.List
var msgBuff chan *utils.Message

var mainGui *gui.MainGui

func main() {
	msgBuff = make(chan *utils.Message, MSG_BUFF_SIZE)
	initFlags()
	flag.Parse()
	validateFlags()

	if logsEnabled {
		logFile, err := os.OpenFile("serial_monitor_logs.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		utils.Must("open log file", err)
		fmt.Printf("Logs will be written to %s\n", logFile.Name())
		defer logFile.Close()
		log.SetOutput(logFile)
	} else {
		log.SetOutput(io.Discard)
	}

	log.Println("Initializing serial monitor")
	logFlags()

	inputMode = false
	followMode = true
	paused = false
	fullScreen = false
	printTime = false
	hexMode = false

	portName = getPort()
	openSerial(portName)
	defer closeSerial()

	gui.Init()
	defer gui.Close()
	createGui()

	messages = list.New()
	go handleMessages()
	go readSerial()
	mainGui.Render()

	var input bytes.Buffer
	clearInputFn := func() {
		input.Reset()
		mainGui.InputParagraph.Text = getInputPrefix()
		mainGui.Render()
	}
	uiEvents := ui.PollEvents()
	for {
		e := <-uiEvents
		if e.Type != ui.KeyboardEvent {
			continue
		}
		if e.ID == "<Escape>" {
			if inputMode {
				inputMode = false
				mainGui.InputParagraph.Text = getInstructions()
				mainGui.Render()
				log.Println("Exiting input mode")
			} else {
				log.Println("Exiting program")
				break
			}
		} else if inputMode {
			if e.ID == "<Backspace>" {
				if input.Len() > 0 {
					input.Truncate(input.Len() - 1)
					mainGui.InputParagraph.Text = getInputPrefix() + input.String()
					mainGui.Render()
				}
			} else {
				if e.ID == "<Enter>" && serialPort != nil {
					n, err := serialPort.Write(input.Bytes())
					utils.Must("write to serial", err)
					writtenBytes += int64(n)
					clearInputFn()
					updateWrittenBytesParagraph()
					mainGui.Render()
				} else {
					input.WriteString(uiEventToChar(e.ID))
					mainGui.InputParagraph.Text = getInputPrefix() + input.String()
					mainGui.Render()
				}
			}
		} else {
			switch e.ID {
			case "i":
				if !paused {
					inputMode = true
					mainGui.InputParagraph.Text = getInputPrefix() + input.String()
					if guiMode == gui.Text && messages.Len() > 0 && followMode {
						mainGui.InboxList.ScrollBottom()
					}
					log.Println("Entering input mode")
					mainGui.Render()
				}
			case "p":
				log.Println("Pausing/Unpausing")
				pauseOrUnpause(portName)
				updatePauseParagraph()
				mainGui.Render()
			case "m":
				changeGuiMode()
				mainGui.Render()
			case "z":
				zoomInOut()
				mainGui.Render()
			case "c":
				if guiMode == gui.Text && messages.Len() > 0 {
					mainGui.InboxList.ScrollTop()
				}
				messages = list.New()
				if guiMode == gui.Text {
					updateMsgInbox()
				} else if guiMode == gui.Plot {
					updatePlot()
				}
				mainGui.Render()
			}
			if guiMode == gui.Text {
				switch e.ID {
				case "j":
					mainGui.InboxList.ScrollHalfPageDown()
					mainGui.Render()
				case "k":
					mainGui.InboxList.ScrollHalfPageUp()
					mainGui.Render()
				case "b":
					mainGui.InboxList.ScrollBottom()
					mainGui.Render()
				case "t":
					mainGui.InboxList.ScrollTop()
					mainGui.Render()
				case "f":
					followMode = !followMode
					updateFollowParagraph()
					mainGui.Render()
				case "s":
					printTime = !printTime
					updateMsgInbox()
					updateTimestampsEnabledParagraph()
					mainGui.Render()
				case "h":
					hexMode = !hexMode
					updateMsgInbox()
					updateHexModeParagraph()
					mainGui.Render()
				}
			}
		}
	}
}

func uiEventToChar(eventId string) string {
	if eventId == "<Space>" {
		return " "
	} else if len(eventId) > 1 {
		return ""
	}
	return eventId
}

func zoomInOut() {
	fullScreen = !fullScreen
	restartGui()
}

func changeGuiMode() {
	if guiMode == gui.Text {
		guiMode = gui.Plot
	} else {
		guiMode = gui.Text
	}
	restartGui()
}

func restartGui() {
	paused = true
	gui.Close()
	gui.Init()
	createGui()
	if guiMode == gui.Text {
		updateMsgInbox()
	} else {
		updatePlot()
	}
	paused = false
	updatePauseParagraph()
}

func handleMessages() {
	for msg := range msgBuff {
		if messages.Len() > MAX_MSG_CAPACITY {
			messages.Remove(messages.Back())
		}
		messages.PushFront(msg)
		if guiMode == gui.Text {
			updateMsgInbox()
		} else if guiMode == gui.Plot {
			updatePlot()
		}
		updateReadBytesParagraph()
		mainGui.Render()
	}
}

func updateHexModeParagraph() {
	if !fullScreen {
		mainGui.HexModeParagraph.Text = fmt.Sprintf("Hexmode: %v", hexMode)
	}
}

func updateTimestampsEnabledParagraph() {
	if !fullScreen {
		mainGui.TimestampsEnabledParagraph.Text = fmt.Sprintf("Timestamps: %v", printTime)
	}
}

func updateFollowParagraph() {
	if !fullScreen {
		mainGui.FollowModeParagraph.Text = fmt.Sprintf("Follow: %v", followMode)
	}
}

func updatePauseParagraph() {
	if !fullScreen {
		mainGui.PauseParagraph.Text = fmt.Sprintf("Pause: %v", paused)
	}
}

func updateWrittenBytesParagraph() {
	if !fullScreen {
		mainGui.WrittenDataParagraph.Text = fmt.Sprintf("Written [B]: %d", writtenBytes)
	}
}

func updateReadBytesParagraph() {
	if !fullScreen {
		mainGui.ReadDataParagraph.Text = fmt.Sprintf("Read [B]: %d", readBytes)
	}
}

func updatePlot() {
	mainGui.InboxPlot.Data = convertMsgsToPoints()
}

func convertMsgsToPoints() [][]float64 {
	points := make([][]float64, 1)
	var pointsCount int
	if messages.Len() > MAX_POINT_CAPACITY {
		pointsCount = MAX_POINT_CAPACITY
	} else {
		pointsCount = messages.Len()
	}
	points[0] = utils.ListToSliceFloat(messages, pointsCount)
	return points
}

func updateMsgInbox() {
	mainGui.InboxList.Rows = utils.ListToSliceMsg(messages, messages.Len(), printTime, hexMode)
	if followMode && messages.Len() > 0 {
		mainGui.InboxList.ScrollBottom()
	}
}

func createGui() {
	log.Printf("Creating gui in %s mode\n", guiMode)
	mainGui = gui.NewMainGui(guiMode, fullScreen)

	if !fullScreen {
		mainGui.BaudParagraph.Text = fmt.Sprintf("Baud: %d", baud)
		mainGui.DeviceParagraph.Text = fmt.Sprintf("Device: %s", portName)
		mainGui.ReadTimeoutParagraph.Text = fmt.Sprintf("Read timeout [ms]: %d", readTimeoutMillieconds)
		mainGui.LogsEnabledParagraph.Text = fmt.Sprintf("Logs enabled: %v", logsEnabled)
	}
	updateWrittenBytesParagraph()
	updateReadBytesParagraph()
	updatePauseParagraph()
	if guiMode == gui.Text {
		updateFollowParagraph()
		updateHexModeParagraph()
		updateTimestampsEnabledParagraph()
	}
	mainGui.InputParagraph.Text = getInstructions()
}

func pauseOrUnpause(portName string) {
	if paused {
		log.Println("Unpausing")
		if serialPort == nil {
			openSerial(portName)
			paused = false
		}
	} else {
		log.Println("Pausing")
		if serialPort != nil {
			paused = true
			closeSerial()
		}
	}
}

func closeSerial() {
	if serialPort == nil {
		log.Println("Serial port already closed")
	}
	utils.Must("drain serial", serialPort.Drain())
	utils.Must("close serial", serialPort.Close())
	serialPort = nil
	log.Println("Serial port closed")
}

func openSerial(portName string) {
	if serialPort != nil {
		log.Fatalln("serial port hasn't been closed in order to be opened")
	}
	var err error
	serialMode := &serial.Mode{BaudRate: baud}
	serialPort, err = serial.Open(portName, serialMode)
	serialPort.SetReadTimeout(time.Duration(int32(readTimeoutMillieconds)) * time.Millisecond)
	utils.Must("open serial", err)
	utils.Must("flush", serialPort.Drain())
	utils.Must("reset input buffer", serialPort.ResetInputBuffer())
	utils.Must("reset output buffer", serialPort.ResetOutputBuffer())
	log.Printf("Serial port to %s opened\n", portName)
}

func getInstructions() string {
	if guiMode == gui.Text {
		return TEXT_NAVIGATION_INSTRUCTIONS
	} else if guiMode == gui.Plot {
		return PLOT_NAVIGATION_INSTRUCTIONS
	} else {
		return ""
	}
}

func getInputPrefix() string {
	if inputMode {
		return INPUT_PREFIX
	} else {
		return ""
	}
}

func getPort() string {
	ports, err := serial.GetPortsList()
	utils.Must("get ports", err)
	if len(ports) == 0 {
		log.Fatalln("no serial ports found!")
	}
	fmt.Printf("Choose one of given ports (type in number 1-%d):\n", len(ports))
	sort.Strings(ports)
	for i, port := range ports {
		fmt.Printf("%d. %s\n", i+1, port)
	}
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	utils.Must("read user input", err)
	chosenPosition, err := strconv.Atoi(line[:len(line)-1])
	utils.Must("chose option", err)
	if chosenPosition < 1 || chosenPosition > len(ports) {
		log.Fatalln("invalid chosen port")
	}
	log.Printf("Chosen port: %s\n", ports[chosenPosition-1])
	return ports[chosenPosition-1]
}

func readSerial() {
	var buff bytes.Buffer
	temp_buff := make([]byte, 512)
	for {
		if paused || serialPort == nil {
			continue
		}
		n, _ := serialPort.Read(temp_buff)
		if n != 0 {
			readBytes += int64(n)
			buff.Write(temp_buff[:n])
		} else {
			if buff.Len() > 0 {
				line, err := buff.ReadString('\n')
				if err != nil {
					buff.Write([]byte(line))
				} else {
					msgBuff <- utils.NowMessage(line)
				}
			}
		}
	}
}

func initFlags() {
	flag.IntVar(&baud, "baud", 9600, "Baud value")
	flag.IntVar(&readTimeoutMillieconds, "read-timeout-ms", 10, "Read timeout in milliseconds")
	flag.StringVar(&guiMode, "mode", "TEXT", "Mode for the gui")
	flag.BoolVar(&logsEnabled, "logs", false, "Is logging enabled?")
}

func validateFlags() {
	if baud < 0 {
		log.Fatalln("baud cannot be negative")
	}
	if readTimeoutMillieconds < 0 {
		log.Fatalln("read timeout seconds cannot be negative")
	}
	validMode := true
	for _, availableMode := range gui.GetAvailableModes() {
		validMode = strings.EqualFold(guiMode, availableMode)
		if validMode {
			break
		}
	}
	if !validMode {
		log.Fatalln("invalid mode")
	}
}

func logFlags() {
	log.Printf("Baud rate: %d\n", baud)
	log.Printf("Read timeout [ms]: %d\n", readTimeoutMillieconds)
	log.Printf("Gui mode: %s\n", guiMode)
	log.Printf("Logs enabled: %v\n", logsEnabled)
}
