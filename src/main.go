package main

import (
	assigner "Project/assignment"
	"Project/config"
	"Project/distribution"
	"Project/elevatorStates"
	"Project/localElevator/door"
	"Project/localElevator/elevator"
	"Project/localElevator/elevio"
	"Project/localElevator/motor"
	"Project/network/peers"
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

	//Channel from elevatorStates to assigner
	ch_elevMap := make(chan map[string]types.Elevator, 1)

	//Channels between distributor and assigner
	ch_peerStatus := make(chan peers.PeerUpdate)
	ch_assignedOrder := make(chan types.AssignedOrder, 1)

	//Channels between localElevator and distributor
	ch_newLocalOrder := make(chan elevio.ButtonEvent, config.NumButtons*config.NumFloors)
	ch_localOrderCompleted := make(chan elevio.ButtonEvent, config.NumButtons)
	ch_cancelOrder := make(chan elevio.ButtonEvent, config.NumElevators)

	//Channel from localElevator to elevatorStates
	ch_localElevatorState := make(chan types.Elevator, 1)

	//Channels between door and localElevator
	ch_openDoor := make(chan bool, 1)
	ch_doorClosed := make(chan bool, 1)

	//Channel from localElevator to motor
	ch_setMotorDirn := make(chan elevio.MotorDirection, 1)

	//Channel from motor and door to distributor
	ch_stuck := make(chan bool, 1)

	go elevio.PollButtons(ch_hwButtonPress)
	go elevio.PollFloorSensor(ch_hwFloor)
	go elevio.PollObstructionSwitch(ch_hwObstruction)

	go elevator.LocalElevator(ch_newLocalOrder, ch_hwFloor, ch_localElevatorState, ch_localOrderCompleted, ch_openDoor, ch_doorClosed, ch_setMotorDirn, ch_cancelOrder)
	go motor.Motor(ch_stuck, ch_setMotorDirn)
	go door.Door(ch_hwObstruction, ch_openDoor, ch_stuck, ch_doorClosed)
	go assigner.Assignment(id, ch_peerStatus, ch_elevMap, ch_hwButtonPress, ch_assignedOrder)
	go distribution.Distribution(id, ch_peerStatus, ch_assignedOrder, ch_newLocalOrder, ch_localOrderCompleted, ch_stuck, ch_cancelOrder)
	go elevatorStates.ElevatorStates(id, ch_localElevatorState, ch_elevMap, ch_newLocalOrder)

	ch_wait := make(chan bool)
	<-ch_wait
}
