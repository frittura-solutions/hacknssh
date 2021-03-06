package main

import (
    "github.com/fatih/color"
)

const (
    red     = color.FgRed
    green   = color.FgGreen
    yellow  = color.FgYellow
    blue    = color.FgBlue
    magenta = color.FgMagenta
    cyan    = color.FgCyan
    white   = color.FgWhite

    // Characters for rendering
    verticalWall   = '║'
    horizontalWall = '═'
    topLeft        = '╔'
    topRight       = '╗'
    bottomRight    = '╝'
    bottomLeft     = '╚'

    grass   = ' '
    blocker = '■'
)

var colors = map[string]color.Attribute{
    "red":     red,
    "green":   green,
    "yellow":  yellow,
    "blue":    blue,
    "magenta": magenta,
    "cyan":    cyan,
    "white":   white,
}

type Widget struct {
    Width   int
    Height  int
    Display    [][]string
}

type MainWidget struct {
    Widget
}

func NewWidget(w int, h int, col string) *Widget {
    display := make([][]string, w+2)
    for x := range display {
        display[x] = make([]string, h+2)
    }

    // Load the walls into the rune slice
    borderColorizer := color.New(colors[col]).SprintFunc()
    for x := 0; x < w+2; x++ {
        display[x][0] = borderColorizer(string(horizontalWall))
        display[x][h+1] = borderColorizer(string(horizontalWall))
    }
    for y := 0; y < h+2; y++ {
        display[0][y] = borderColorizer(string(verticalWall))
        display[w+1][y] = borderColorizer(string(verticalWall))
    }

    // Load the walls into the rune slice
    for x := 0; x < w+2; x++ {
        display[x][0] = borderColorizer(string(horizontalWall))
        display[x][h+1] = borderColorizer(string(horizontalWall))
    }
    for y := 0; y < h+2; y++ {
        display[0][y] = borderColorizer(string(verticalWall))
        display[w+1][y] = borderColorizer(string(verticalWall))
    }

    // Time for the edges!
    display[0][0] = borderColorizer(string(topLeft))
    display[w+1][0] = borderColorizer(string(topRight))
    display[w+1][h+1] = borderColorizer(string(bottomRight))
    display[0][h+1] = borderColorizer(string(bottomLeft))

    return &Widget{
        Width:  w,
        Height: h,
        Display: display,
    }
}

func (w *Widget) writeAtLine(str string, h int, align string, pad int, col string) {
    colorStrColorizer := color.New(colors[col]).SprintFunc()
    if align == "right" {
        for i, r := range str {
            charsRemaining := len(str) - i
            if w.Width-pad-charsRemaining >= 0 {
                w.Display[w.Width-pad-charsRemaining][h] = colorStrColorizer(string(r))
            } else {
                break
            }
            
        }
    } else {
        for i, r := range str {
            if pad + i < w.Width {
                w.Display[pad+i][h] = colorStrColorizer(string(r))
            } else {
                break
            }
            
        } 
    }
    
}

func (w *Widget) writeField(player *Player) {
    // Load the players into the rune slice
    pos := player.Pos
    colorizer := color.New(colors[player.Color]).SprintFunc()
    w.Display[pos.RoundX()+1][pos.RoundY()+1] = colorizer(string(player.Marker))

    // Load the player's trail into the rune slice
    for _, segment := range player.Trail {
        x, y := segment.Pos.RoundX()+1, segment.Pos.RoundY()+1
        colorizer := color.New(colors[segment.Color]).SprintFunc()
        w.Display[x][y] = colorizer(string(segment.Marker))
    }
}