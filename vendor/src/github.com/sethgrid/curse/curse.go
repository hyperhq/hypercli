package curse

// http://en.wikipedia.org/wiki/ANSI_escape_code#Sequence_elements

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/kless/term"
)

type Cursor struct {
	Position
	StartingPosition Position
	Style

	terminal *term.Terminal
}

type Position struct {
	X, Y int
}

type Style struct {
	Foreground, Background, Bold int
}

func New() (*Cursor, error) {
	col, line, err := GetCursorPosition()
	if err != nil {
		return &Cursor{}, err
	}

	c := &Cursor{}
	c.Position.X, c.StartingPosition.X = col, col
	c.Position.Y, c.StartingPosition.Y = line, line
	c.terminal, err = term.New()
	return c, err
}

func (c *Cursor) MoveUp(nLines int) *Cursor {
	fmt.Printf("%c[%dA", ESC, nLines)
	c.Position.Y -= nLines
	return c
}

func (c *Cursor) MoveDown(nLines int) *Cursor {
	fmt.Printf("%c[%dB", ESC, nLines)
	c.Position.Y += nLines
	return c
}

func (c *Cursor) MoveRight(nSpaces int) *Cursor {
	c.Position.X += nSpaces
	c.Move(c.Position.X, c.Position.Y)
	return c
}

func (c *Cursor) MoveLeft(nSpaces int) *Cursor {
	c.Position.X -= nSpaces
	c.Move(c.Position.X, c.Position.Y)
	return c
}

func (c *Cursor) EraseCurrentLine() *Cursor {
	fmt.Printf("%c[2K\r", ESC)
	c.Position.X = 1
	return c
}

func (c *Cursor) EraseUp() *Cursor {
	fmt.Printf("%c[1J", ESC)
	return c
}

func (c *Cursor) EraseDown() *Cursor {
	fmt.Printf("%c[0J", ESC)
	return c
}

func (c *Cursor) EraseAll() *Cursor {
	fmt.Printf("%c[0J", ESC)
	return c
}

func (c *Cursor) Reset() *Cursor {
	c.Move(c.StartingPosition.X, c.StartingPosition.Y)
	return c
}

func (c *Cursor) Move(col, line int) *Cursor {
	fmt.Printf("%c[%d;%df", ESC, line, col)
	c.Position.X = col
	c.Position.Y = line
	return c
}

func (c *Cursor) SetColor(color int) *Cursor {
	fmt.Printf("%c[%dm", ESC, FORGROUND+color)
	c.Style.Foreground = color
	c.Style.Bold = 0
	return c
}

func (c *Cursor) SetColorBold(color int) *Cursor {
	fmt.Printf("%c[%d;1m", ESC, FORGROUND+color)
	c.Style.Foreground = color
	c.Style.Bold = 1
	return c
}

func (c *Cursor) SetBackgroundColor(color int) *Cursor {
	fmt.Printf("%c[%dm", ESC, BACKGROUND+color)
	c.Style.Foreground = color
	c.Style.Bold = 0
	return c
}

func (c *Cursor) SetDefaultStyle() *Cursor {
	fmt.Printf("%c[39;49m", ESC)
	c.Style.Foreground = 0
	c.Style.Bold = 0
	return c
}

func (c *Cursor) ModeRaw() *Cursor {
	_ = c.terminal.RawMode()

	return c
}

func (c *Cursor) ModeRestore() *Cursor {
	_ = c.terminal.Restore()

	return c
}

// using named returns to help when using the method to know what is what
func GetScreenDimensions() (cols int, lines int, err error) {
	// todo: use kless/term to listen in on screen size changes
	// get size
	cmd := exec.Command("/bin/stty", "size")
	cmd.Stdin = os.Stdin
	size, err := cmd.Output()
	if err != nil {
		return 70, 15, errors.New(fmt.Sprintf("unable to get dimensions - %s", err))
	}
	parts := strings.Split(strings.TrimSpace(string(size)), " ")

	if len(parts) != 2 {
		return 70, 15, errors.New(fmt.Sprintf("unable to parse dimensions - %s", err))
	}

	// make ints
	cols, err = strconv.Atoi(parts[1])
	if err != nil {
		return cols, 15, errors.New(fmt.Sprintf("unable to get int dimensions - %s", err))
	}
	lines, err = strconv.Atoi(parts[0])
	if err != nil {
		return cols, 15, errors.New(fmt.Sprintf("unable to get int dimensions - %s", err))
	}

	return cols, lines, nil
}

func fallback_SetRawMode() {
	rawMode := exec.Command("/bin/stty", "raw")
	rawMode.Stdin = os.Stdin
	_ = rawMode.Run()
	rawMode.Wait()
}

func fallback_SetCookedMode() {
	// I've noticed that this does not always work when called from
	// inside the program. From command line, you can run the following
	// '$ go run calling_app.go; stty -raw'
	// if you lose the ability to visably enter new text
	cookedMode := exec.Command("/bin/stty", "-raw")
	cookedMode.Stdin = os.Stdin
	_ = cookedMode.Run()
	cookedMode.Wait()
}

func GetCursorPosition() (col int, line int, err error) {
	// set terminal to raw mode and back
	t, err := term.New()
	if err != nil {
		fallback_SetRawMode()
		defer fallback_SetCookedMode()
	} else {
		t.RawMode()
		defer t.Restore()
	}

	// same as $ echo -e "\033[6n"
	// by printing the output, we are triggering input
	fmt.Printf(fmt.Sprintf("\r%c[6n", ESC))

	// capture keyboard output from print command
	reader := bufio.NewReader(os.Stdin)

	// capture the triggered stdin from the print
	text, _ := reader.ReadSlice('R')

	// check for the desired output
	re := regexp.MustCompile(`\d+;\d+`)
	res := re.FindString(string(text))

	// make sure that cooked mode gets set
	if res != "" {
		parts := strings.Split(res, ";")
		line, _ = strconv.Atoi(parts[0])
		col, _ = strconv.Atoi(parts[1])
		return col, line, nil

	} else {
		return 0, 0, errors.New("unable to read cursor position")
	}
}

const (
	// control
	ESC = 27

	// style
	BLACK   = 0
	RED     = 1
	GREEN   = 2
	YELLOW  = 3
	BLUE    = 4
	MAGENTA = 5
	CYAN    = 6
	WHITE   = 7

	FORGROUND  = 30
	BACKGROUND = 40
)
