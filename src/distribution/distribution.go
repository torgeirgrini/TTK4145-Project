package distribution

import (
	"Project/config"
	"Project/localElevator/elevator"
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
	orderState OrderState
	ackList    [config.NumElevators]string
}

type ElevatorStateMessage struct {
	ID            string
	LocalElevator elevator.Elevator
	HallCalls     [][]HallCall
	ElevStateMap  map[string]elevator.Elevator
}

//Denne har ansvar for ordrebestillinger, og håndtere ordretap osv
//Sende ourOrders til localElevator
//Send allOrders for å kunne sette lys overalt for eksempel

/*
func Distribution(id string,
	ch_newLocalOrder <-chan elevio.ButtonType,
	ch_NewElevatorStateMap chan<- map[string]elevator.Elevator,
	ch_newLocalElevator <-chan elevator.Elevator) {

	elevatorSystemMap := make(map[string]elevator.Elevator)
	for {
		select {

		case newOrder := <-ch_newLocalOrder:
			elevatorSystemMap = updateMapWithOrder(id, elevatorSystemMap, newOrder)
			ch_unassignedOrder <- elevatorSystemMap     //Til assigner
			ch_NewElevatorStateMap <- elevatorSystemMap //Til network

		case updatedElevatorSystemMap := <-ch_updatedElevatorSystemMap:
			//sebd den ut på nettet!!
			//send vår elevstruct til localElevator fsm

		case newLocalElevator := <-ch_newLocalElevator:
			elevatorSystemMap = updateMapWithLocalElevator(id, elevatorSystemMap, newLocalElevator)
			// case <-ch_doorTimer:

			// case obstruction = <-ch_Obstruction:

		}
	}

}
*/

func DeepCopy(elevators map[string]elevator.Elevator) map[string]elevator.Elevator {
	copied := make(map[string]elevator.Elevator)
	for i, e := range elevators {
		copied[i] = e
	}
	return copied
}

func Distribution(
	localID string,
	localElevator <-chan elevator.Elevator,
	allElevators chan map[string]elevator.Elevator,
	tx chan<- ElevatorStateMessage,
	rx <-chan ElevatorStateMessage,
	ch_peerUpdate <-chan peers.PeerUpdate) {

	elevators := make(map[string]elevator.Elevator)

	tick := time.NewTicker(config.TransmitInterval * time.Millisecond)

	Hallcalls := make([][]HallCall, config.NumFloors)
	for i := range Hallcalls {
		Hallcalls[i] = make([]HallCall, config.NumButtons-1)
	}
	for {
		select {
		case e := <-localElevator:
			if !reflect.DeepEqual(elevators[localID], e) {
				elevators[localID] = e

				allElevators <- DeepCopy(elevators)

			}
		case <-tick.C:

			tx <- ElevatorStateMessage{localID, elevators[localID], Hallcalls, elevators}

		case remote := <-rx:
			if !reflect.DeepEqual(elevators[remote.ID], remote.LocalElevator) {
				elevators[remote.ID] = remote.LocalElevator
				allElevators <- DeepCopy(elevators)
				fmt.Println("ELEVATOR STATE MAP:", remote.ElevStateMap)
			}
		case peerUpdate := <-ch_peerUpdate:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", peerUpdate.Peers)
			fmt.Printf("  New:      %q\n", peerUpdate.New)
			fmt.Printf("  Lost:     %q\n", peerUpdate.Lost)
			//Må si ifra om at noen har kommet på/fallt av nettet
			//Kan for eksmpel gjøres ved å sette available bit i elevators(ESM'en)

		}
	}
}

