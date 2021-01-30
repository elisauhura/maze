package levels

const level0 string = `mazemap
4 1
dddb
0
0 0 r
3 0
`

const level1 string = `mazemap
4 1
dddb
1
2 0 g r
0 0 r
3 0
`

const level2 string = `mazemap
8 4
xoorxxxx
xoorxxxx
xdobxxxx
dddddddb
2
0 3 b r
2 2 g d
2 1 d
7 3
`

var Levels = [...]Level{
	{"level 0", level0},
	{"level 1", level1},
	{"level 2", level2},
}

type Level struct {
	Name string
	Data string
}

/*
Maze map format
mazemap
WIDTH HEIGHT
xodrb.. HEIGTH LINES of WIDTH chars + \n
... x -> no space
... o -> no walls
... d -> wall on the down side
... r -> wall on the right side
... b -> wall on both side
NUMBER OF ELEMENTS
X Y ELEMENT DIRECTION... // Element may be
... g -> movable wall
... r -> right turn monster
... l -> left turn monster
... b -> bounce back monster
... // Direction
... udlr <- keyboard
STARTX STARTY DIRECTION
ENDX ENDY
*/
