package board

import (
	"encoding/json"
	"github.com/stianeikeland/go-rpio"
	"speedPicker/config"
	"time"
)

type Lock struct {
	Pin          rpio.Pin      `json:"-"`
	Name         string        `json:"human_name"`
	SolvedState   rpio.State   `json:"-"`
	PickDuration time.Duration `json:"pick_duration"`
	DebounceTimeSeconds time.Duration `json:"-"`
	MonitorCtl   chan int      `json:"-"`
}

type Board struct {
	Locks       []Lock    `json:"locks"`
	StartButton rpio.Pin  `json:"-"`
	ResetButton rpio.Pin  `json:"-"`
	StatusLed   rpio.Pin  `json:"-"`
	StartedAt   time.Time `json:"start_time"`
	Running     bool      `json:"running"`
	ButtonMonitorCtl        chan int  `json:"-"`
}

// New creates and returns an instance of a Board, initialized with the provided configuration
func New(config *config.BoardConfig) *Board {
	newBoard := &Board{}
	newBoard.Init(config)

	return newBoard
}

// Init initializes a Board instance with a given configuration
func (b *Board) Init(config *config.BoardConfig) {
	if err := rpio.Open(); err != nil {
		panic(err)
	}

	b.ButtonMonitorCtl = make(chan int, 1)

	b.StartButton = rpio.Pin(config.StartButtonPin)
	b.StartButton.Input()

	b.ResetButton = rpio.Pin(config.ResetButtonPin)
	b.ResetButton.Input()

	b.StatusLed = rpio.Pin(config.StatusLedPin)
	b.StatusLed.Output()
	b.StatusLed.Write(rpio.Low)

	b.Locks = make([]Lock, len(config.Locks))

	for i := 0; i < len(config.Locks); i++ {
		b.Locks[i].Pin = rpio.Pin(config.Locks[i].Pin)
		b.Locks[i].Pin.Input()
		b.Locks[i].SolvedState = rpio.State(config.Locks[i].SolvedState)
		b.Locks[i].DebounceTimeSeconds = time.Duration(config.LockDebounceTimeSeconds) * time.Second
		b.Locks[i].Name = config.Locks[i].Name
		b.Locks[i].MonitorCtl = make(chan int, 1)
	}
}

// Start sets a Board to its running state, and starts monitor goroutines for
// each Lock in the Board
func (b *Board) Run() {
	startBtnPush := make(chan int, 1)
	go monitorPin(b.StartButton, rpio.High, true, 1 * time.Second, startBtnPush, b.ButtonMonitorCtl)

	resetBtnPush := make(chan int, 1)
	go monitorPin(b.ResetButton, rpio.High, true, 1 * time.Second, resetBtnPush, b.ButtonMonitorCtl)

	for {
		select {
		case <-startBtnPush:
			go b.Start()
		case <-resetBtnPush:
			go b.Reset()
		}
	}
}

// Start sets a Board to its running state, and starts monitor goroutines for
// each Lock in the Board
func (b *Board) Start() {
	if b.Running == true {
		return
	}


	b.StartedAt = time.Now()
	b.StatusLed.Write(rpio.High)
	b.Running = true

	for i := 0; i < len(b.Locks); i++ {
		go b.Locks[i].Monitor()
	}

}

// Stop sets a Board to its idle state, and signals each Lock's monitoring
// goroutine to exit
func (b *Board) Stop() {
	if b.Running == false {
		return
	}


	b.Reset()
	b.ButtonMonitorCtl <- 0
	b.ButtonMonitorCtl <- 0

	if err := rpio.Close(); err != nil {
		panic(err)
	}
}

// Reset stops a running or idle Board, and overwrites any time
// data recorded for the Board itself as well as each Lock
func (b *Board) Reset() {

	// Kill all lock monitors
	for i := 0; i < len(b.Locks); i++ {
		if b.Locks[i].PickDuration == 0 {
			b.Locks[i].MonitorCtl <- 0
		}
	}

	b.StatusLed.Write(rpio.Low)
	b.StartedAt = time.Time{}
	b.Running = false

	for i := 0; i < len(b.Locks); i++ {
		b.Locks[i].PickDuration = 0
	}
}

// Monitor watches the state of a Lock for changes, and records the
// current timestamp when one has been detected.
func (l *Lock) Monitor() {
	started := time.Now()

	stateChg := make(chan int, 1)
	monCtl := make(chan int, 1)
	go monitorPin(l.Pin, l.SolvedState, true, l.DebounceTimeSeconds, stateChg, monCtl)

	for {
		select {
		case <-l.MonitorCtl:
			monCtl <- 0
		case pinState := <-stateChg:
			if pinState == 1 {
				l.PickDuration = time.Since(started)
			}
			return
		}
	}
}

// MonitorPin watches the state of a Pin for changes, and signals on a
// channel when one has been detected.
// Pin Monitors will send `0` on their control channel if they have received a
// termination signal. Otherwise, if a change in state is observed and recorded,
// a Pin Monitor will send `1` over its control channel before exiting.
// Each Pin Monitor's channel may also be used by a parent goroutine
// for control. If a Monitor receives any value over this channel, it will
// terminate early.
func monitorPin(pin rpio.Pin, alertState rpio.State, shouldDebounce bool, coolDownPeriod time.Duration, state chan int, ctl chan int) {
	lastTrig := time.Now().Add(-coolDownPeriod)

	for {
		select {
		case <-ctl:
			ctl <- 0
			return
		default:
			if debounce(shouldDebounce, coolDownPeriod, lastTrig) && pin.Read() == alertState {
				lastTrig = time.Now()
				state <- 1
			}

			// Sample time is each 10ms
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func debounce(shouldDebounce bool, coolDownPeriod time.Duration, lastTriggered time.Time) (bool) {
	return shouldDebounce && time.Since(lastTriggered) >= coolDownPeriod
}

// String enables a Board to be serialized as a JSON string
func (b *Board) String() string {
	bytes, err := json.Marshal(b)
	if err != nil {
		return ""
	}

	return string(bytes)
}
