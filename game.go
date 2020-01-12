package main

import (
	"bufio"
	"fmt"
	
	"golang.org/x/crypto/ssh"
	
)

type ByColor []*Player

func (slice ByColor) Len() int {
	return len(slice)
}

// func (slice ByColor) Less(i, j int) bool {
// 	return playerColors[slice[i].Color] < playerColors[slice[j].Color]
// }

func (slice ByColor) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

type TileType int

const (
	TileGrass TileType = iota
	TileBlocker
)

type Tile struct {
	Type TileType
}

const (
	gameWidth  = 84
	gameHeight = 36

	keyW = 'w'
	keyA = 'a'
	keyS = 's'
	keyD = 'd'

	keyZ = 'z'
	keyQ = 'q'
	// keyS and keyD are already defined

	keyH = 'h'
	keyJ = 'j'
	keyK = 'k'
	keyL = 'l'

	keyComma = ','
	keyO     = 'o'
	keyE     = 'e'

	keyCtrlC  = 3
	keyEscape = 27
)

type GameManager struct {
	Rooms         map[string]*Room
	HandleChannel chan ssh.Channel
}

func NewGameManager() *GameManager {
	return &GameManager{
		Rooms:         map[string]*Room{},
		HandleChannel: make(chan ssh.Channel),
	}
}

// getRoomWithAvailability returns a reference to a game with available spots for
// players. If one does not exist, nil is returned.
func (gm *GameManager) getRoomWithAvailability() *Room {
	var g *Room

	for _, game := range gm.Rooms {
		spots := game.maxPlayers - game.SessionCount()
		if spots > 0 {
			g = game
			break
		}
	}

	return g
}

func (gm *GameManager) SessionCount() int {
	sum := 0
	for _, game := range gm.Rooms {
		sum += game.SessionCount()
	}
	return sum
}

func (gm *GameManager) RoomCount() int {
	return len(gm.Rooms)
}

func (gm *GameManager) HandleNewChannel(c ssh.Channel, name string) {
	g := gm.getRoomWithAvailability()
	if g == nil {
		g = NewRoom(gameWidth, gameHeight)
		gm.Rooms[g.Name] = g

		go g.Run()
	}

	colorOptions := g.AvailableColors()
	finalColor := colorOptions[0]


	session := NewSession(c, g.WorldWidth(), g.WorldHeight(), name, finalColor)
	g.AddSession(session)

	go func() {
		reader := bufio.NewReader(c)
		for {
			r, _, err := reader.ReadRune()
			if err != nil {
				fmt.Println(err)
				break
			}

			switch r {
			case keyW:
				session.Player.HandleUp()
			case keyA:
				session.Player.HandleLeft()
			case keyS:
				session.Player.HandleDown()
			case keyD:
				session.Player.HandleRight()
			case keyL:
				session.Player.HandleAccelerate()
			case keyK:
				session.Player.HandleBreak()
			case keyCtrlC, keyEscape:
				if g.SessionCount() == 1 {
					delete(gm.Rooms, g.Name)
				}

				g.RemoveSession(session)
			}
		}
	}()
	fmt.Printf("Player joined. Current stats: %d users, %d rooms\n",
			gm.SessionCount(), gm.RoomCount())
}

//, keyZ, keyK, keyComma, keyQ, keyH, keyJ, keyO, keyL, keyE
