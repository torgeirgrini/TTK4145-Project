package main

import (
	assigner "Project/assignment"
	"Project/config"
	"Project/distribution"
	"Project/localElevator/elevator"
	"Project/localElevator/elevio"
	"Project/localElevator/fsm"
	"Project/network/bcast"
	"Project/network/peers"
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
	allElevators := make(chan map[string]elevator.Elevator)

	//Network channels
	tx := make(chan distribution.ElevatorStateMessage)
	rx := make(chan distribution.ElevatorStateMessage)

	ch_peerTxEnable := make(chan bool)
	ch_peerUpdate := make(chan peers.PeerUpdate)

	//Local elevator state channel
	ch_localElevatorStruct := make(chan elevator.Elevator)
	//ch_localOrder = make(chan elevator.Elevator)

	go elevio.PollButtons(ch_drv_buttons)
	go elevio.PollFloorSensor(ch_drv_floors)
	go elevio.PollObstructionSwitch(ch_drv_obstr)

	go fsm.RunElevator(ch_drv_buttons, ch_drv_floors, ch_drv_obstr, ch_localElevatorStruct)

	go bcast.Transmitter(config.PortBroadcast, tx)
	go bcast.Receiver(config.PortBroadcast, rx)

	go peers.Transmitter(config.PortPeers, id, ch_peerTxEnable)
	go peers.Receiver(config.PortPeers, ch_peerUpdate)

	go assigner.Assignment(allElevators)
	go distribution.Distribution(id, ch_localElevatorStruct, allElevators, tx, rx, ch_peerUpdate)

	ch_wait := make(chan bool)
	<-ch_wait
}
