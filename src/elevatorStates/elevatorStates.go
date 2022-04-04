package elevatorStates

import (
	"Project/config"
	"Project/localElevator/elevio"
	"Project/network/bcast"
	"Project/network/peers"
	"Project/types"
	"Project/utilities"
	"reflect"
	"time"
)

func ElevatorStates(
	localID 					   string,
	ch_localElevatorState <-chan   types.Elevator,
	ch_elevMap 				chan<- map[string]types.Elevator,
	ch_newLocalOrder 		chan<- elevio.ButtonEvent,
) {
	tick := time.NewTicker(config.TransmitInterval_ms * time.Millisecond)

	ch_txElevator := make(chan types.ElevStateNetMsg, config.NumElevators) 
	ch_rxElevator := make(chan types.ElevStateNetMsg, config.NumElevators)
	ch_peerUpdate := make(chan peers.PeerUpdate, config.NumElevators)
	ch_peerTxEnable := make(chan bool)
	ch_tick := tick.C

	elevators := make(map[string]types.Elevator)

	go bcast.Transmitter(config.PortBroadcast, ch_txElevator)
	go bcast.Receiver(config.PortBroadcast, ch_rxElevator)
	go peers.Transmitter(config.PortPeersElevStates, localID, ch_peerTxEnable)
	go peers.Receiver(config.PortPeersElevStates, ch_peerUpdate)

	peerAvailability := peers.PeerUpdate{
		Peers: []string{localID},
		New:   "",
		Lost:  make([]string, 0),
	}

	elevators[localID] = <-ch_localElevatorState

	InitTimer := time.NewTimer(time.Duration(config.InitTime_s) * time.Second)
	ch_initTimer := InitTimer.C
	init := true
	for init {
		select {
		case <-ch_initTimer:
			init = false
		case initState := <-ch_rxElevator:
			if initState.ElevStateID == localID {
				for floor := 0; floor < config.NumFloors; floor++ {
					if initState.ElevState.Requests[floor][elevio.BT_Cab] {
						ch_newLocalOrder <- elevio.ButtonEvent{
							Floor:  floor,
							Button: elevio.BT_Cab,
						}
					}
				}
				init = false
			}
		case e := <-ch_localElevatorState:
			elevators[localID] = e
		default:
		}
	}
	ch_elevMap <- utilities.DeepCopyElevatorMap(elevators)
	for {
		select {
		case e := <-ch_localElevatorState:
			if !reflect.DeepEqual(elevators[localID], e) {
				elevators[localID] = utilities.DeepCopyElevatorStruct(e)
				ch_elevMap <- utilities.DeepCopyElevatorMap(elevators)
			}
		case <-ch_tick:
			ch_txElevator <- types.ElevStateNetMsg{
				SenderID:    localID,
				ElevStateID: localID,
				ElevState:   utilities.DeepCopyElevatorStruct(elevators[localID]),
			}
		case remote := <-ch_rxElevator:
			if remote.SenderID != localID {
				if !reflect.DeepEqual(elevators[remote.ElevStateID], remote.ElevState) && remote.ElevStateID == remote.SenderID {
					elevators[remote.ElevStateID] = utilities.DeepCopyElevatorStruct(remote.ElevState)
					ch_elevMap <- utilities.DeepCopyElevatorMap(elevators)
				}
			}
		case peerAvailability = <-ch_peerUpdate:
			if peerAvailability.New != localID && peerAvailability.New != "" {
				if _, ok := elevators[peerAvailability.New]; ok {
					ch_txElevator <- types.ElevStateNetMsg{
						SenderID:    localID,
						ElevStateID: peerAvailability.New,
						ElevState:   utilities.DeepCopyElevatorStruct(elevators[peerAvailability.New]),
					}
				}
			}
		}
	}
}
