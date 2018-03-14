package shared

type Coord struct {
	X float64
	Y float64
}

type RegistrationDetails struct {
	Connections []string
	Identifier int
	InitState	InitialState
}

type InitialState struct {
	Settings EnvironmentSettings
	CatchWorth	int
}

type EnvironmentSettings struct {
	WinMaxX float64
	WinMaxY float64
	WallCoords []Coord
}