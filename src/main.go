package main

import (
	assigner "Project/assignment"
	"Project/config"
	"Project/distribution"
	"Project/localElevator/elevio"
	"Project/localElevator/fsm"
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
	ch_drv_buttons := make(chan elevio.ButtonEvent)
	ch_drv_floors := make(chan int)
	ch_drv_obstr := make(chan bool)

	//Assigner channels
	allElevators := make(chan map[string]types.Elevator, 1)
	ch_orderAssigned := make(chan types.MsgToDistributor, 1)

	//Network channels

	ch_peerTxEnable := make(chan bool)

	//Local elevator state channel
	ch_localElevatorStruct := make(chan types.Elevator, 1)
	ch_newLocalOrder := make(chan elevio.ButtonEvent, 1)

	go elevio.PollButtons(ch_drv_buttons)
	go elevio.PollFloorSensor(ch_drv_floors)
	go elevio.PollObstructionSwitch(ch_drv_obstr)

	go fsm.RunElevator(ch_newLocalOrder, ch_drv_floors, ch_drv_obstr, ch_localElevatorStruct)

	go peers.Transmitter(config.PortPeers, id, ch_peerTxEnable)

	go assigner.Assignment(id, allElevators, ch_drv_buttons, ch_orderAssigned)
	go distribution.Distribution(id, ch_localElevatorStruct, allElevators, ch_orderAssigned, ch_newLocalOrder)

	ch_wait := make(chan bool)
	<-ch_wait
}
