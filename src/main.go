package main

import (
	assigner "Project/assignment"
	"Project/config"
	"Project/distribution"
	"Project/localElevator/door"
	"Project/localElevator/elevio"
	"Project/localElevator/fsm"
	"Project/localElevator/motor"
	"Project/types"
	"flag"
)

func main() {
	var id string
	var p string
	flag.StringVar(&p, "p", "15657", "port to elevator server")
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()

	elevio.Init("localhost:"+p, config.NumFloors)

	//Hardware channels
	ch_hwButtonPress := make(chan elevio.ButtonEvent)
	ch_hwFloor := make(chan int)
	ch_hwObstruction := make(chan bool)

	//Assigner<-/->Distributor channels
	ch_informationToAssigner := make(chan types.AssignerMessage, 1)
	ch_assignedOrder := make(chan types.AssignedOrder, 1)

	//Network channels

	//LocalElevator<-/->Distributor channels
	ch_newLocalOrder := make(chan elevio.ButtonEvent, config.NumButtons*config.NumFloors)
	ch_localOrderCompleted := make(chan elevio.ButtonEvent, 2)
	ch_localElevatorState := make(chan types.Elevator, 1)
	//ch_loneElevator := make(chan bool)

	//Door channels
	ch_openDoor := make(chan bool, 1)
	ch_doorClosed := make(chan bool, 1)

	//Motor channels
	ch_setMotorDirn := make(chan elevio.MotorDirection, 1)

	//Channel for stuckness
	ch_stuck := make(chan bool, 1)

	go elevio.PollButtons(ch_hwButtonPress)
	go elevio.PollFloorSensor(ch_hwFloor)
	go elevio.PollObstructionSwitch(ch_hwObstruction)

	go fsm.RunLocalElevator(ch_newLocalOrder, ch_hwFloor, ch_localElevatorState, ch_localOrderCompleted, ch_openDoor, ch_doorClosed, ch_stuck, ch_setMotorDirn)
	go motor.Motor(ch_stuck, ch_setMotorDirn)
	go door.Door(ch_hwObstruction, ch_openDoor, ch_stuck, ch_doorClosed)
	go assigner.Assignment(id, ch_informationToAssigner, ch_hwButtonPress, ch_assignedOrder)
	go distribution.Distribution(id, ch_localElevatorState, ch_informationToAssigner, ch_assignedOrder, ch_newLocalOrder, ch_localOrderCompleted)

	ch_wait := make(chan bool)
	<-ch_wait
}
