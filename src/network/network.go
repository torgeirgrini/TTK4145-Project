package network

import (
	"Project/config"
	"Project/localElevator/elevator"
	"Project/network/bcast"
	"Project/network/peers"
	"flag"
	"fmt"
)

//Kanskje dele opp i to funskjoner?
func Network(
	id string,
	ch_TxNewElevatorStateMap chan<- map[string]elevator.Elevator,
	ch_RxNewElevatorStateMap <-chan map[string]elevator.Elevator,
	ch_peerTxEnable <-chan bool) {

	// Our id can be anything. Here we pass it on the command line, using
	//  `go run main.go -id=our_id`

	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()

	ch_peerUpdate := make(chan peers.PeerUpdate)

	go peers.Transmitter(config.PortPeers, id, ch_peerTxEnable)
	go peers.Receiver(config.PortPeers, ch_peerUpdate)

	go bcast.Transmitter(config.PortBroadcast, ch_TxNewElevatorStateMap)
	go bcast.Receiver(config.PortBroadcast, ch_RxNewElevatorStateMap)

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
			//inn på en ny channel
		}
	}
}
