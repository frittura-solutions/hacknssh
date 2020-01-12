package main

import (
    "time"

    "golang.org/x/crypto/ssh"
)

type Session struct {
    c ssh.Channel

    LastAction time.Time
    HighScore  int
    Player     *Player
    Display    string
}

func NewSession(c ssh.Channel, worldWidth, worldHeight int, name string, color string) *Session {

    s := Session{c: c, LastAction: time.Now()}
    s.newRoom(worldWidth, worldHeight, name, color)

    return &s
}

func (s *Session) newRoom(worldWidth, worldHeight int, name string, color string) {
    s.Player = NewPlayer(s, worldWidth, worldHeight, name, color)
    
}

func (s *Session) didAction() {
    s.LastAction = time.Now()
}

func (s *Session) StartOver(worldWidth, worldHeight int) {
    s.newRoom(worldWidth, worldHeight, s.Player.Name, s.Player.Color)
}

func (s *Session) Read(p []byte) (int, error) {
    return s.c.Read(p)
}

func (s *Session) Write(p []byte) (int, error) {
    return s.c.Write(p)
}
