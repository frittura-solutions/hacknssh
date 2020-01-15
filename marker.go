package main

const(
    UpRune    = '⇡'
    LeftRune  = '⇠'
    DownRune  = '⇣'
    RightRune = '⇢'

    TrailHorizontal      = '┄'
    TrailVertical        = '┆'
    TrailLeftCornerUp    = '╭'
    TrailLeftCornerDown  = '╰'
    TrailRightCornerDown = '╯'
    TrailRightCornerUp   = '╮'
)


type Marker struct{
    Up              rune
    Left            rune
    Down            rune
    Right           rune

    Horizontal      rune
    Vertical        rune
    LeftUpCorner    rune
    LeftDownCorner    rune
    RightUpCorner    rune
    RightDownCorner    rune
}

func NewMarker() Marker {
    return Marker{
        Up:     UpRune,       
        Left:   LeftRune,           
        Down:   DownRune,
        Right:  RightRune,           

        Horizontal: TrailHorizontal,      
        Vertical:   TrailVertical,
        LeftUpCorner:   TrailLeftCornerUp,    
        LeftDownCorner:   TrailLeftCornerDown,        
        RightUpCorner:   TrailRightCornerUp,        
        RightDownCorner:   TrailRightCornerDown,        
    }
}