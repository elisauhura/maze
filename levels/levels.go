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
3 3
oor
oor
ddb
4
1 0 g d
1 1 g l
1 1 g r
1 1 g d
0 1 r
1 1
`

const level3 string = `mazemap
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

const level4 string = `mazemap
9 3
dodrxodrx
xrxodrxob
xddbxddbx
2
2 0 r r
6 2 l r
0 0 r
8 1
`

const level5 string = `mazemap
2 7
rx
rx
rx
or
rr
rr
db
2
0 1 b d
0 3 g d
1 6
0 0
`

var Levels = [...]Level{
	{"level 0", level0},
	{"level 1", level1},
	{"level 2", level2},
	{"level 3", level3},
	{"level 4", level4},
	{"level 5", level5},
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
