package main

import (
    "time"
    "math/rand"
    "math"
)

type Position struct {
    X float64
    Y float64
}

func PositionFromInt(x, y int) Position {
    return Position{float64(x), float64(y)}
}

func (p Position) RoundX() int {
    return int(p.X + 0.5)
}

func (p Position) RoundY() int {
    return int(p.Y + 0.5)
}

type PlayerDirection int

const (
    verticalPlayerSpeed        = 0.007
    horizontalPlayerSpeed      = 0.01
    playerCountScoreMultiplier = 1.25
    playerTimeout              = 15 * time.Second

    playerUpRune    = '⇡'
    playerLeftRune  = '⇠'
    playerDownRune  = '⇣'
    playerRightRune = '⇢'

    playerTrailHorizontal      = '┄'
    playerTrailVertical        = '┆'
    playerTrailLeftCornerUp    = '╭'
    playerTrailLeftCornerDown  = '╰'
    playerTrailRightCornerDown = '╯'
    playerTrailRightCornerUp   = '╮'

    PlayerUp PlayerDirection = iota
    PlayerLeft
    PlayerDown
    PlayerRight
)

var playerColors = []string{
    "red", "green", "yellow", "blue",
    "magenta", "cyan", "white",
}

type PlayerTrailSegment struct {
    Marker rune
    Pos    Position
    Color  string
}

type Player struct {
    s *Session

    Name      string
    CreatedAt time.Time
    Direction PlayerDirection
    Marker    rune
    Color     string
    Pos       *Position
    Speed     float64

    Trail []PlayerTrailSegment

    score float64
}

// NewPlayer creates a new player. If color is below 1, a random color is chosen
func NewPlayer(s *Session, worldWidth, worldHeight int, name string, color string) *Player {

    rand.Seed(time.Now().UnixNano())

    startX := rand.Float64() * float64(worldWidth)
    startY := rand.Float64() * float64(worldHeight)

    if color == "random" {
        color = playerColors[rand.Intn(len(playerColors))]
    }

    return &Player{
        s:         s,
        Name: name,
        CreatedAt: time.Now(),
        Marker:    playerDownRune,
        Direction: PlayerDown,
        Color:     color,
        Pos:       &Position{startX, startY},
        Speed: 0.5 + rand.Float64(),
    }
}

func (p *Player) addTrailSegment(pos Position, marker rune) {
    segment := PlayerTrailSegment{marker, pos, p.Color}
    p.Trail = append([]PlayerTrailSegment{segment}, p.Trail...)
}

func (p *Player) calculateScore(delta float64, playerCount int) float64 {
    rawIncrement := (delta * (float64(playerCount-1) * playerCountScoreMultiplier))

    // Convert millisecond increment to seconds
    actualIncrement := rawIncrement / 1000

    return p.score + actualIncrement
}

func (p *Player) HandleUp() {
    if p.Direction == PlayerDown {
        return
    }
    p.Direction = PlayerUp
    p.Marker = playerUpRune
    p.s.didAction()
}

func (p *Player) HandleLeft() {
    if p.Direction == PlayerRight {
        return
    }
    p.Direction = PlayerLeft
    p.Marker = playerLeftRune
    p.s.didAction()
}

func (p *Player) HandleDown() {
    if p.Direction == PlayerUp {
        return
    }
    p.Direction = PlayerDown
    p.Marker = playerDownRune
    p.s.didAction()
}

func (p *Player) HandleRight() {
    if p.Direction == PlayerLeft {
        return
    }
    p.Direction = PlayerRight
    p.Marker = playerRightRune
    p.s.didAction()
}

func (p *Player) HandleAccelerate() {
    p.Speed = math.Min(2, p.Speed + 0.25)
    p.s.didAction()
}

func (p *Player) HandleBreak() {
    p.Speed = math.Max(0, p.Speed - 0.25)
    p.s.didAction()
}

func (p *Player) Score() int {
    return int(p.score)
}

func (p *Player) Update(g *Room, delta float64) {
    startX, startY := p.Pos.RoundX(), p.Pos.RoundY()

    switch p.Direction {
    case PlayerUp:
        p.Pos.Y -= verticalPlayerSpeed * delta * p.Speed
    case PlayerLeft:
        p.Pos.X -= horizontalPlayerSpeed * delta * p.Speed
    case PlayerDown:
        p.Pos.Y += verticalPlayerSpeed * delta * p.Speed
    case PlayerRight:
        p.Pos.X += horizontalPlayerSpeed * delta * p.Speed
    }

    endX, endY := p.Pos.RoundX(), p.Pos.RoundY()

    // If we moved, add a trail segment.
    if endX != startX || endY != startY {
        var lastSeg *PlayerTrailSegment
        var lastSegX, lastSegY int
        if len(p.Trail) > 0 {
            lastSeg = &p.Trail[0]
            lastSegX = lastSeg.Pos.RoundX()
            lastSegY = lastSeg.Pos.RoundY()
        }

        pos := PositionFromInt(startX, startY)

        switch {
        // Handle corners. This took an ungodly amount of time to figure out. Highly
        // recommend you don't touch.
        case lastSeg != nil &&
            (p.Direction == PlayerRight && endX > lastSegX && endY < lastSegY) ||
            (p.Direction == PlayerDown && endX < lastSegX && endY > lastSegY):
            p.addTrailSegment(pos, playerTrailLeftCornerUp)
        case lastSeg != nil &&
            (p.Direction == PlayerUp && endX > lastSegX && endY < lastSegY) ||
            (p.Direction == PlayerLeft && endX < lastSegX && endY > lastSegY):
            p.addTrailSegment(pos, playerTrailRightCornerDown)
        case lastSeg != nil &&
            (p.Direction == PlayerDown && endX > lastSegX && endY > lastSegY) ||
            (p.Direction == PlayerLeft && endX < lastSegX && endY < lastSegY):
            p.addTrailSegment(pos, playerTrailRightCornerUp)
        case lastSeg != nil &&
            (p.Direction == PlayerRight && endX > lastSegX && endY > lastSegY) ||
            (p.Direction == PlayerUp && endX < lastSegX && endY < lastSegY):
            p.addTrailSegment(pos, playerTrailLeftCornerDown)

        // Vertical and horizontal trails
        case endX == startX && endY < startY:
            p.addTrailSegment(pos, playerTrailVertical)
        case endX < startX && endY == startY:
            p.addTrailSegment(pos, playerTrailHorizontal)
        case endX == startX && endY > startY:
            p.addTrailSegment(pos, playerTrailVertical)
        case endX > startX && endY == startY:
            p.addTrailSegment(pos, playerTrailHorizontal)
        }
    }

    p.score = p.calculateScore(delta, len(g.players()))
}