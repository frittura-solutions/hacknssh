package main

import (
    "fmt"
    "time"

    "github.com/dustinkirkland/golang-petname"
    "github.com/PuerkitoBio/goquery"
)

func hasHref(index int, element *goquery.Selection) bool {
    // See if the href attribute exists on the element
    _, exists := element.Attr("href")
    if exists {
        return true
    }
    return false
}

func getHref(index int, element *goquery.Selection) string {
    // See if the href attribute exists on the element
    href, exists := element.Attr("href")
    if exists {
        return href
    }
    return ""
}

func getLinks(document *goquery.Document) []string {
    links := document.Find("a").FilterFunction(hasHref).Map(getHref)//  
    fmt.Printf("%v", links)    
    return links
}


type Room struct {
    Name        string
    Width       int
    Height      int
    Redraw      chan struct{}
    HighScore   int
    maxPlayers  int
    Page        string
    Links       []string
    MainWidget  *Widget
    SideWidget  *Widget//move to Session


    // Top left is 0,0
    level [][]Tile
    hub   Hub
}

func NewRoom(worldWidth, worldHeight int) *Room {
    //defaultUrl := "https://en.wikipedia.org/w/api.php?action=query&titles=Albert%20Einstein&prop=links&plnamespace=0&pllimit=500"
    // defaultUrl, err := url.Parse("https://en.wikipedia.org/w/api.php")
    // if err != nil {
       
    // }
    startPage := "Albert Einstein"
    // parameters := url.Values{}
    // parameters.Add("action", "query")
    // parameters.Add("titles", startPage)
    // parameters.Add("prop", "links")
    // parameters.Add("plnamespace", "0")
    // parameters.Add("pllimit", "500")
    // defaultUrl.RawQuery = parameters.Encode()

    // fmt.Printf("Encoded URL is %q\n", defaultUrl.String())
    // resp, err := http.Get(defaultUrl.String())
    // if err != nil {
    //     // handle error
    // }
    // defer resp.Body.Close()

    r := &Room{
        Name:   petname.Generate(1, ""),
        Redraw: make(chan struct{}),
        maxPlayers: 8,
        hub:    NewHub(),
        Page: startPage,
    }
    r.initalize(worldWidth, worldHeight)

    return r
}

func (r *Room) initalize(width, height int) {
    r.level = make([][]Tile, width)
    for x := range r.level {
        r.level[x] = make([]Tile, height)
    }

    // Default world to grass
    for x := range r.level {
        for y := range r.level[x] {
            r.setTileType(Position{float64(x), float64(y)}, TileGrass)
        }
    }
    //r.setTileType(Position{float64(r.WorldWidth()/2), float64(r.WorldHeight()/2)}, TileBlocker)
    
}

func (r *Room) setTileType(pos Position, tileType TileType) error {
    outOfBoundsErr := "The given %s value (%s) is out of bounds"
    if pos.RoundX() > len(r.level) || pos.RoundX() < 0 {
        return fmt.Errorf(outOfBoundsErr, "X", pos.X)
    } else if pos.RoundY() > len(r.level[pos.RoundX()]) || pos.RoundY() < 0 {
        return fmt.Errorf(outOfBoundsErr, "Y", pos.Y)
    }

    r.level[pos.RoundX()][pos.RoundY()].Type = tileType

    return nil
}

func (r *Room) players() map[*Player]*Session {
    players := make(map[*Player]*Session)

    for session := range r.hub.Sessions {
        players[session.Player] = session
    }

    return players
}



func (r *Room) WorldWidth() int {
    return len(r.level)
}

func (r *Room) WorldHeight() int {
    return len(r.level[0])
}

func (r *Room) AvailableColors() []string {
    usedColors := map[string]bool{}
    for _, color := range playerColors {
        usedColors[color] = false
    }

    for player := range r.players() {
        usedColors[player.Color] = true
    }

    availableColors := []string{}
    for color, used := range usedColors {
        if !used {
            availableColors = append(availableColors, color)
        }
    }

    return availableColors
}

func (r *Room) SessionCount() int {
    return len(r.hub.Sessions)
}

func (r *Room) Run() {
    // Proxy r.Redraw's channel to r.hub.Redraw
    go func() {
        for {
            r.hub.Redraw <- <-r.Redraw
        }
    }()

    // Run game loop
    go func() {
        var lastUpdate time.Time

        c := time.Tick(time.Second / 60)
        for now := range c {
            r.Update(float64(now.Sub(lastUpdate)) / float64(time.Millisecond))

            lastUpdate = now
        }
    }()

    // Redraw regularly.
    //
    // TODO: Implement diffing and only redraw when needed
    go func() {
        c := time.Tick(time.Second / 10)
        for range c {
            r.Redraw <- struct{}{}
        }
    }()

    r.hub.Run(r)
}

// Update is the main game logic loop. Delta is the time since the last update
// in milliseconds.
func (r *Room) Update(delta float64) {
    // We'll use this to make a set of all of the coordinates that are occupied by
    // trails
    trailCoordMap := make(map[string]*PlayerTrailSegment)

    // Update player data
    for player, session := range r.players() {
        player.Update(r, delta)

        // Update session high score, if applicable
        if player.Score() > session.HighScore {
            session.HighScore = player.Score()
        }

        // Update global high score, if applicable
        if player.Score() > r.HighScore {
            r.HighScore = player.Score()
        }

        // Restart the player if they're out of bounds
        pos := player.Pos
        if pos.RoundX() < 0 || pos.RoundX() >= len(r.level) ||
            pos.RoundY() < 0 || pos.RoundY() >= len(r.level[0]) {
            session.StartOver(r.WorldWidth(), r.WorldHeight())
        }

        // Kick the player if they've timed out
        if time.Now().Sub(session.LastAction) > playerTimeout {
            fmt.Fprint(session, "\r\n\r\nYou were terminated due to inactivity\r\n")
            r.RemoveSession(session)
            return
        }

        // range gives copies, but we need a reference in the trailCoordMap so we
        // can modify the Color value if there is a collision, so iterate by index
        // instead.
        for i := range player.Trail {
            seg := &player.Trail[i]
            coordStr := fmt.Sprintf("%d,%d", seg.Pos.RoundX(), seg.Pos.RoundY())
            trailCoordMap[coordStr] = seg
        }
    }

    // Check if any players collide with a trail and restart them if so
    for player, session := range r.players() {
        playerPos := fmt.Sprintf("%d,%d", player.Pos.RoundX(), player.Pos.RoundY())
        if segment, collided := trailCoordMap[playerPos]; collided {
            segment.Color = player.Color
            session.StartOver(r.WorldWidth(), r.WorldHeight())
        }
    }
}

// func (r *Room) Render(s *Session) {
//     worldStr := r.worldString(s)

//     var b bytes.Buffer
//     b.WriteString("\033[H\033[2J")
//     b.WriteString(worldStr)

//     // Send over the rendered world
//     io.Copy(s, &b)
// }

func (r *Room) AddSession(s *Session) {
    r.hub.Register <- s
}

func (r *Room) RemoveSession(s *Session) {
    r.hub.Unregister <- s
}

