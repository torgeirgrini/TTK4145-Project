package main

import (
	"Project/config"
	"Project/localElevator/elevio"
	"Project/localElevator/fsm"
	"Project/localElevator/timer"
	//"fmt"
)

func main() {

	elevio.Init("localhost:15657", config.NumFloors)

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)

	//timer channel:
	ch_timerTimedOut := make(chan bool)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)
	go timer.PollTimerTimedOut(ch_timerTimedOut)

	//run FSM:
	fsm.RunElevator(drv_buttons, drv_floors, ch_timerTimedOut, drv_obstr)

}
