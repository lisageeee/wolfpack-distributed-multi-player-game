package shared

type Coord struct {
	X float64
	Y float64
}

type RegistrationDetails struct {
	Connections []string
	Identifier int
}

type InitialState struct {
	NumPoints	int
	CatchWorth	int
}