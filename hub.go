package main

import (
    "fmt"
)

type Hub struct {
    Sessions   map[*Session]struct{}
    Redraw     chan struct{}
    Register   chan *Session
    Unregister chan *Session
}

func NewHub() Hub {
    return Hub{
        Sessions:   make(map[*Session]struct{}),
        Redraw:     make(chan struct{}),
        Register:   make(chan *Session),
        Unregister: make(chan *Session),
    }
}

func (h *Hub) Run(r *Room) {
    for {
        select {
        case <-h.Redraw:
            for s := range h.Sessions {
                go s.Render(r)
            }
        case s := <-h.Register:
            // Hide the cursor
            fmt.Fprint(s, "\033[?25l")

            h.Sessions[s] = struct{}{}
        case s := <-h.Unregister:
            if _, ok := h.Sessions[s]; ok {
                fmt.Fprint(s, "\r\n\r\n~ End of Line ~ \r\n\r\nRemember to use WASD to move!\r\n\r\n")

                // Unhide the cursor
                fmt.Fprint(s, "\033[?25h")

                delete(h.Sessions, s)
                s.c.Close()
            }
        }
    }
}