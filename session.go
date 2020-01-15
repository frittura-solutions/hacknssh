package main

import (
    "io"
    "bytes"
    "sort"
    "fmt"
    "time"

    "golang.org/x/crypto/ssh"
)

type Session struct {
    c ssh.Channel

    LastAction time.Time
    HighScore  int
    Player     *Player
    Display    string
    MainWidget  *Widget
    SideWidgets  []*Widget
}

func NewSession(c ssh.Channel, worldWidth, worldHeight int, name string, color string) *Session {
    s := Session{c: c, LastAction: time.Now()}
    s.Player = NewPlayer(&s, worldWidth, worldHeight, name, color)//insert here logic to instead assign already existing player
    s.MainWidget = NewWidget(worldWidth, worldHeight, "white")
    return &s
}

func (s *Session) didAction() {
    s.LastAction = time.Now()
}

func (s *Session) StartOver(worldWidth, worldHeight int) {
    s.Player = NewPlayer(s, worldWidth, worldHeight, s.Player.Name, s.Player.Color)
}

func (s *Session) Read(p []byte) (int, error) {
    return s.c.Read(p)
}

func (s *Session) Write(p []byte) (int, error) {
    return s.c.Write(p)
}

// Warning: this will only work with square worlds
func (s *Session) worldString(r *Room) string {
    panel := NewWidget(32, s.MainWidget.Height, "white")
    // Create side panel
    for x := 0; x < panel.Width; x++ {
        for y := 0; y < panel.Height; y++ {
            panel.Display[x+1][y+1] = string(grass)
        }
    }

    // Draw the player's score
    scoreStr := fmt.Sprintf(
        " %s Score: %d : Your High Score: %d : Room High Score: %d ",
        s.Player.Name,
        s.Player.Score(),
        s.HighScore,
        r.HighScore,
    )
    s.MainWidget.writeAtLine(scoreStr, 0, "left", 3, "white")
    // Draw the room's name
    nameStr := fmt.Sprintf(" %s ", r.Name)
    s.MainWidget.writeAtLine(nameStr, 0, "right", 3, "white")


    // Draw everyone's scores
    if len(r.players()) > 1 {
        // Sort the players by color name
        players := []*Player{}

        for player := range r.players() {
            
            players = append(players, player)
        }

        sort.Sort(ByColor(players))

        // Actually draw their scores
        for i, player := range players {
            scoreStr := fmt.Sprintf(" %s: %d",
                player.Name,
                player.Score(),
            )
            panel.writeAtLine(scoreStr, i+1, "left", 1, player.Color)
        }
    } else {
        warning := " All alone in this lonely room "
        panel.writeAtLine(warning, 1, "left", 1, "white")
    }

    

    // Load the level into the string slice
    for x := 0; x < s.MainWidget.Width; x++ {
        for y := 0; y < s.MainWidget.Height; y++ {
            tile := r.level[x][y]

            switch tile.Type {
            case TileGrass:
                s.MainWidget.Display[x+1][y+1] = string(grass)
            case TileBlocker:
                s.MainWidget.Display[x+1][y+1] = string(blocker)
            }
        }
    }

    

    // Draw the player's color

    //playerSpeedStr := fmt.Sprintf("S: %3.2f ", s.Player.Speed)
    

    // for i, r := range playerSpeedStr {
    //     charsRemaining := len(playerSpeedStr) - i
    //     strPanel[len(strPanel)-3-charsRemaining][2] = colorStrColorizer(string(r))
    // }

    // Load the players into the rune slice
    // for player := range r.players() {

    //     pos := player.Pos
    //     //colorizer := color.New(player.Color).SprintFunc()
    //     s.MainWidget.Display[pos.RoundX()+1][pos.RoundY()+1] = colorizer(string(player.Marker))

    //     // Load the player's trail into the rune slice
    //     for _, segment := range player.Trail {
    //         x, y := segment.Pos.RoundX()+1, segment.Pos.RoundY()+1
    //         colorizer := color.New(segment.Color).SprintFunc()
    //         s.MainWidget.Display[x][y] = colorizer(string(segment.Marker))
    //     }
    // }
    for player := range r.players() {
        s.MainWidget.writeField(player)
    }

    // Convert the rune slice to a string
    totalWidth := s.MainWidget.Width//+panelWidth
    totalHeight := s.MainWidget.Height//+panelHeight
    buffer := bytes.NewBuffer(make([]byte, 0, totalWidth*totalHeight*2))
    for y := 0; y < s.MainWidget.Height+2; y++ {
        
        for x := 0; x < panel.Width+2; x++ {
            buffer.WriteString(panel.Display[x][y])
        }
        for x := 0; x < s.MainWidget.Width+2; x++ {
            buffer.WriteString(s.MainWidget.Display[x][y])
        }
        // Don't add an extra newline if we're on the last iteration
        if y != s.MainWidget.Height+2-1 {
            buffer.WriteString("\r\n")
        }
    }

    return buffer.String()
}

func (s *Session) Render(r *Room) {
    worldStr := s.worldString(r)

    var b bytes.Buffer
    b.WriteString("\033[H\033[2J")
    b.WriteString(worldStr)

    // Send over the rendered world
    io.Copy(s, &b)
}
