package distributor

import (
	"Project/localElevator/elevator"
	"Project/localElevator/elevio"
)


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
func updateMapWithOrder(id string, elevatorSystemMap map[string]elevator.Elevator, newOrder elevio.ButtonType) map[string]elevator.Elevator {
	//oppdater elevatorSystemMap[id]
	return elevatorSystemMap
}

func updateMapWithLocalElevator(id string, elevatorSystemMap map[string]elevator.Elevator, newLocalElevator elevator.Elevator) map[string]elevator.Elevator {
	//oppdater elevatorSystemMap[id]
	return elevatorSystemMap
}
