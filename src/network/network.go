package network

import (
	"Project/config"
	"Project/localElevator/elevator"
	"Project/localElevator/elevio"
	"Project/network/bcast"
	"Project/network/peers"
	"fmt"
	"reflect"
	"time"
)


type OrderState int

const (
	OS_NONE OrderState = iota
	OS_UNCONFIRMED 
	OS_CONFIRMED
	OS_COMPLETED
)

type HallCall struct {
	executerID string
	assignerID string
	orderState config.OrderState
	ackList    [config.NumElevators]string
}

type ElevatorStateMessage struct {
	ID string
	E  elevator.Elevator
	HallCalls [][]HallCall
}

func States(
	localID string,
	localElevator <-chan elevator.Elevator,
	allElevators chan<- map[string]elevator.Elevator,
) {

/*TODO:
	Må oppdatere peer list
	Må håndtere å sende/motta hallcalls
*/
	elevators := make(map[string]elevator.Elevator)

	tx := make(chan ElevatorStateMessage)
	rx := make(chan ElevatorStateMessage)

	go bcast.Transmitter(config.PortBroadcast, tx)
	go bcast.Receiver(config.PortBroadcast, rx)

	tick := time.NewTicker(config.TransmitInterval * time.Millisecond)

	copy := func(map[string]elevator.Elevator) map[string]elevator.Elevator {
		copied := make(map[string]elevator.Elevator)
		for i, e := range elevators {
			copied[i] = e
		}
		return copied
	}

	//Gå gjennom alle elevators, ore alle hallcalls
	Hallcalls := func (map[string]elevator.Elevator) [][]HallCall {
		var HallCalls [][]HallCall
		for j := 0; j < config.NumButtons-1; j++ {
			for i, e := range elevators {
				if e.Requests[i][j] {

				}
				}	
				
				HallCalls[i][j] = HallCalls[i][j] || e.Requests[i][j]
			}
		}
	
	}



	for {
		select {
		case e := <-localElevator:
			if !reflect.DeepEqual(elevators[localID], e) {
				elevators[localID] = e
				allElevators <- copy(elevators)
			}
		case <-tick.C:
			tx <- ElevatorStateMessage{localID, elevators[localID], Hallcalls}
		case remote := <-rx:
			if !reflect.DeepEqual(elevators[remote.ID], remote.E) {
				elevators[remote.ID] = remote.E
				allElevators <- copy(elevators)
			}
		}
	}
}

////////////////////////////////////////////////

func Receive(
	ch_RxNewElevatorStateMap <-chan MessageStruct) {

}

//Kanskje dele opp i to funskjoner?
func Network(
	id string,
	ch_TxNewElevatorStateMap chan MessageStruct,
	ch_RxNewElevatorStateMap chan MessageStruct,
	ch_newLocalState <-chan elevator.Elevator,
	ch_peerTxEnable <-chan bool) {

	// Our id can be anything. Here we pass it on the command line, using
	//  `go run main.go -id=our_id`

	ch_peerUpdate := make(chan peers.PeerUpdate)

	go peers.Transmitter(config.PortPeers, id, ch_peerTxEnable)
	go peers.Receiver(config.PortPeers, ch_peerUpdate)

	go bcast.Transmitter(config.PortBroadcast, ch_TxNewElevatorStateMap)
	go bcast.Receiver(config.PortBroadcast, ch_RxNewElevatorStateMap)

	//go PeriodicTransmit(ch_TxNewElevatorStateMap)
	go Receive(ch_RxNewElevatorStateMap)
	fmt.Println("Started")
	//FSM for å motta fra nettet
	for {
		select {
		case peerUpdate := <-ch_peerUpdate:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", peerUpdate.Peers)
			fmt.Printf("  New:      %q\n", peerUpdate.New)
			fmt.Printf("  Lost:     %q\n", peerUpdate.Lost)
			//Må si ifra om at noen har kommet på/fallt av nettet
		case NewElevatorStateMap := <-ch_RxNewElevatorStateMap:
			_ = NewElevatorStateMap
			//inn på en ny channel
		}
	}
}
