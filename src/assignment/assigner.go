package assigner

// import (
// 	"encoding/json"
// 	"fmt"
// 	"os/exec"
// 	"runtime"
// )

// // Struct members must be public in order to be accessible by json.Marshal/.Unmarshal
// // This means they must start with a capital letter, so we need to use field renaming struct tags to make them camelCase

//Denne tar inn 
	//Buttonpress fra hardware IO
	//Den får periodisk inn ESM fra distributor(main akkurat nå)
	//Peer list for å vite hvilke heiser som er på nettet og som kan ta ordre (Denne burde kanskje sendes periodisk)






// type HRAElevState struct {
// 	Behavior    string `json:"behaviour"`
// 	Floor       int    `json:"floor"`
// 	Direction   string `json:"direction"`
// 	CabRequests []bool `json:"cabRequests"`
// }

// type HRAInput struct {
// 	HallRequests [][2]bool               `json:"hallRequests"`
// 	States       map[string]HRAElevState `json:"states"`
// }

// func main() {

// 	hraExecutable := ""
// 	switch runtime.GOOS {
// 	case "linux":
// 		hraExecutable = "hall_request_assigner"
// 	case "windows":
// 		hraExecutable = "hall_request_assigner.exe"
// 	default:
// 		panic("OS not supported")
// 	}

// 	//input := convertToHRAInput(alle heiser, hallrequest);

// 	input := HRAInput{
// 		HallRequests: [][2]bool{{false, false}, {true, false}, {false, false}, {false, true}},
// 		States: map[string]HRAElevState{
// 			"one": HRAElevState{
// 				Behavior:    "moving",
// 				Floor:       2,
// 				Direction:   "up",
// 				CabRequests: []bool{false, false, false, true},
// 			},
// 			"two": HRAElevState{
// 				Behavior:    "idle",
// 				Floor:       0,
// 				Direction:   "stop",
// 				CabRequests: []bool{false, false, false, false},
// 			},
// 		},
// 	}

// 	jsonBytes, err := json.Marshal(input)
// 	fmt.Println("json.Marshal error: ", err)

// 	ret, err := exec.Command("cost_func/"+hraExecutable, "-i", string(jsonBytes)).Output()
// 	fmt.Println("exec.Command error: ", err)

// 	output := new(map[string][][2]bool)
// 	err = json.Unmarshal(ret, &output)
// 	fmt.Println("json.Unmarshal error: ", err)

// 	fmt.Printf("output: \n")
// 	for k, v := range *output {
// 		fmt.Printf("%6v :  %+v\n", k, v)
// 	}

// 	// output := convertToElevators(output_HRAOutput)
// 	// ch_calculated <- output
// }
