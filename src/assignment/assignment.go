package assigner

import (
	"Project/assignment/costfn"
	"Project/config"
	"Project/localElevator/elevio"
	"Project/types"
	"Project/utilities"
	"fmt"
)

// // Struct members must be public in order to be accessible by json.Marshal/.Unmarshal
// // This means they must start with a capital letter, so we need to use field renaming struct tags to make them camelCase

//Denne tar inn
//Buttonpress fra hardware IO
//Den får periodisk inn ESM fra distributor(main akkurat nå)
//Peer list for å vite hvilke heiser som er på nettet og som kan ta ordre (Denne burde kanskje sendes periodisk)

func Assignment(
	localID string,
	ch_informationToAssigner <-chan types.AssignerMessage,
	ch_hwButtonPress <-chan elevio.ButtonEvent,
	ch_assignedOrder chan<- types.AssignedOrder,
) {

	assignerMsg := types.AssignerMessage{}
	elevatorMap := make(map[string]types.Elevator)
	var peerList []string
	//fmt.Printf("Assign elevators: %p, %#+v\n", elevatorMap, elevatorMap)
	//btn_event := elevio.ButtonEvent{}

	for {

		select {
		case assignerMsg = <-ch_informationToAssigner:
			elevatorMap = utilities.DeepCopyElevatorMap(assignerMsg.ElevatorMap) 
			peerList = utilities.DeepCopyStringSlice(assignerMsg.PeerList,len(assignerMsg.PeerList)) 
			//fmt.Printf("assign | new elevators: %+v\n", elevatorMap)
			//fmt.Printf("Assign elevators (after copy): %p, %#+v\n", elevatorMap, elevatorMap)

			//Videresend til assigner sånn at den kan regne ut
		case btn_event := <-ch_hwButtonPress:
			//Here we need to calcilate the cost functin
			if btn_event.Button == elevio.BT_Cab {
				ch_assignedOrder <- types.AssignedOrder{
					OrderType: btn_event,
					ID:        localID,
				}

			} else {

				AssignedElevID := localID

				fmt.Println("Peerlist: ", peerList)
				fmt.Println("Buttonevent")
				fmt.Printf("assign | elevators | %+#v\n", elevatorMap)
				elev_copy := utilities.DeepCopyElevatorStruct(elevatorMap[AssignedElevID])
				fmt.Printf("assign | elev copy | %+#v\n", elev_copy)
				elev_copy.Requests[btn_event.Floor][btn_event.Button] = true
				min_time := costfn.TimeToIdle(elev_copy)

				for _, id := range peerList {
					fmt.Println("ID:", id)
					elev_copy = utilities.DeepCopyElevatorStruct(elevatorMap[id])
					elev_copy.Requests[btn_event.Floor][btn_event.Button] = true
					if costfn.TimeToIdle(elev_copy) < min_time {
						AssignedElevID = id
						min_time = costfn.TimeToIdle(elev_copy)
					}
				}
				fmt.Println("Assigned elevator: ", AssignedElevID)

				ch_assignedOrder <- types.AssignedOrder{OrderType: btn_event, ID: AssignedElevID}
			}
		}
	}
}

func HallRequestsFromESM(allElevators map[string]types.Elevator) [][]bool {
	//var HallCalls [][]HallCall
	Hallcalls := make([][]bool, config.NumFloors)
	for i := 0; i < config.NumFloors; i++ {
		Hallcalls[i] = make([]bool, config.NumButtons-1)
		for j := 0; j < config.NumButtons-1; j++ {
			Hallcalls[i][j] = false
			for _, e := range allElevators {
				Hallcalls[i][j] = Hallcalls[i][j] || e.Requests[i][j]
			}
		}
	}
	return Hallcalls
}
