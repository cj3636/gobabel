package address

type Address struct {
	Version  string
	Segments []string
	Start    *int
	End      *int
}

const Version = "bf1"
