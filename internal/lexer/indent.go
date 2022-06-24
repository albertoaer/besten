package lexer

import (
	"bufio"
	"errors"
	"io"
	"strings"
)

type dirtyBlock struct {
	raw      string
	children []dirtyBlock
	parent   *dirtyBlock
}

var separators string = " \t"

//remain: the remained indent, relative: the indent relative, err: the error
func indentSolver(arr []string, indent string) (remain string, relative int, err error) {
	remain = indent
	relative = 0
	for i, a := range arr {
		if !strings.HasPrefix(remain, a) {
			if len(remain) > 0 {
				err = errors.New("No matching identation level")
			}
			relative = i - len(arr)
			break
		}
		remain = remain[len(a):]
	}
	if len(remain) > 0 {
		relative += 1
	}
	return
}

func getRawStructure(src io.Reader) (blocks []dirtyBlock, err error) {
	reader := bufio.NewReader(src)
	line := ""
	var indent_array []string

	active := &dirtyBlock{"", blocks, nil}
	root := active

	for {
		//Solve raw line
		buffer, notended, e := reader.ReadLine()
		if e != nil {
			if buffer != nil {
				err = e
				return //Error return, reading error
			} else {
				break
			}
		}
		line += string(buffer)
		line = strings.TrimRight(line, separators)
		if notended || len(line) == 0 {
			continue //In the next iteration will read the rest of the line and operate it
		}

		//Solve indentation
		noindent := strings.TrimLeft(line, separators)
		indentsize := len(line) - len(noindent)
		sub_indent, relative, e := indentSolver(indent_array, line[:indentsize])
		line = ""
		if e != nil {
			err = e
			return //Error return, indent error
		}
		if relative < 0 {
			indent_array = indent_array[:len(indent_array)+relative]
		} else if relative > 0 { //Can only be one
			indent_array = append(indent_array, sub_indent)
		}

		//Solve line block
		if relative < 0 {
			if active.parent == nil {
				err = errors.New("Negative block indentation")
				return //Negative block indentation should never occurs, hard internal error
			}
			count := relative
			for count < 0 {
				active = active.parent
				count++
			}
		} else if relative > 0 {
			if len(active.children) == 0 {
				err = errors.New("No block to include line")
				return //First children sets the indentation, so this should never occurss
			}
			active = &active.children[len(active.children)-1]
		}
		active.children = append(active.children, dirtyBlock{noindent, make([]dirtyBlock, 0), active})
	}
	blocks = root.children
	return
}
