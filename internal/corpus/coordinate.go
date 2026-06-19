package corpus

import "strings"

type Coordinate struct {
	Hexagon string `json:"hexagon"`
	Wall    string `json:"wall"`
	Shelf   string `json:"shelf"`
	Book    string `json:"book"`
	Page    string `json:"page"`
}

func (c Coordinate) Normalize() string {
	return strings.Join([]string{c.Hexagon, c.Wall, c.Shelf, c.Book, c.Page}, "/")
}
