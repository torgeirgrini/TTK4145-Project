package distributor

import (
	"Project/localElevator/elevator"
	"Project/localElevator/elevio"
)

func Distribution(ch_newLocalOrder chan elevio.ButtonType) {
	elevatorSystemMap := make(map[string]elevator.Elevator)
	for {
		select {

		case newOrder := <-ch_newLocalOrder:
			elevatorSystemMap = updateMapWithOrder(elevatorSystemMap, newOrder)
			hallcalls := activeHallCalls(elevatorSystemMap, newOrder)
			ch_unassignedOrder <- hallcalls

			// case newFloor := <-ch_FloorArrival:

			// case <-ch_doorTimer:

			// case obstruction = <-ch_Obstruction:

		}
	}

}

func updateMapWithOrder(id string,
						elevatorSystemMap map[string]elevator.Elevator,
						newOrder elevio.ButtonType) {



}
