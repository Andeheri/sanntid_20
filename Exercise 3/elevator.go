package main
import "fmt"

type role string

// Master-slave states
const (
	master  role = "master"
	slave   role = "slave"
	unknown role = "unknown"
)

func setElevatorRole(elevator_role *role){
	*elevator_role = master
}

func main() {
	var elevator_role role = unknown

	setElevatorRole(&elevator_role)

	fmt.Printf("Elevator role: %s\n", elevator_role)
}
