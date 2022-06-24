package lexer

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Source interface {
	Origin() string
	GetSource() (io.ReadCloser, error)
}

type Lexer struct {
	source Source
}

func LexerFor(source Source) *Lexer {
	return &Lexer{source: source}
}

type Block struct {
	Begin    int
	End      int
	Tokens   []Token
	Children []Block
	Parent   *Block
	Origin   string
}

func getChilds(blks []dirtyBlock) (line []Token, sub bool, children []dirtyBlock, lline int, err error) {
	sub = false
	for i := range blks {
		//Read and append the token to the main line
		tk, s, e := GetTokens(blks[i].raw)
		lline = blks[i].line
		if e != nil {
			err = e
			return
		}
		line = append(line, tk...)

		//Check subs or children
		if s || len(blks[i].children) > 0 {
			if i == len(blks)-1 {
				sub = s
				children = blks[i].children
			} else {
				if s {
					err = errors.New("Unexpected block opening")
				} else { //len(blks[i].children) > 0
					err = errors.New("Unexpected indentation")
				}
				return
			}
		} else if i == len(blks)-1 {
			children = make([]dirtyBlock, 0)
		}
	}
	return
}

func (l *Lexer) solveBlock(raw dirtyBlock, parent *Block) (block Block, linenum int, err error) {
	//sublevel: Indicates that any subline is in a sublevel, otherwise sublines would be treat like the same line
	tks, sublevel, err := GetTokens(raw.raw)
	begin := raw.line
	linenum = begin
	end := raw.line
	if err != nil {
		return
	}

	//Multiline and pick children
	target_children := raw.children
	if !sublevel && len(raw.children) > 0 {
		l, sub, target, nline, e := getChilds(target_children)
		if nline > end {
			linenum = nline
			end = nline
		}
		if e != nil {
			err = e
			return
		}
		target_children = target
		sublevel = sub
		tks = append(tks, l...)
	}
	if sublevel != (len(target_children) > 0) {
		err = errors.New("No correspondence between indentation and block opening")
		return
	}

	//Solve children and fill the block template
	childs, ln, e := l.solveBlocks(target_children, &block)
	if e != nil {
		linenum = ln
		err = e
		return
	}
	block.Begin = begin
	block.End = end
	block.Parent = parent
	block.Tokens = tks
	block.Children = childs
	block.Origin = l.source.Origin()
	return
}

func (l *Lexer) solveBlocks(raw []dirtyBlock, parent *Block) (blocks []Block, linenum int, err error) {
	for _, r := range raw {
		b, ln, e := l.solveBlock(r, parent)
		if e != nil {
			linenum = ln
			err = e
			return
		}
		if len(b.Tokens) > 0 {
			blocks = append(blocks, b)
		}
	}
	return
}

func (l *Lexer) GetBlocks() (blocks []Block, err error) {
	var target_line int = -1
	defer func() {
		if err != nil {
			file := l.source.Origin()
			if wdir, err := os.Getwd(); err == nil {
				if relpath, err := filepath.Rel(wdir, file); err == nil {
					file = relpath
				}
			}
			line := ""
			if target_line >= 0 {
				line = fmt.Sprintf(" [Error in line (%d)]", target_line)
			}
			err = fmt.Errorf("[File: %s]%s\n\t%s", file, line, err.Error())
		}
	}()
	var s io.ReadCloser
	if s, err = l.source.GetSource(); err != nil {
		return
	}
	var raw_blocks []dirtyBlock
	raw_blocks, target_line, err = getRawStructure(s)
	if err != nil {
		return
	}
	blocks, target_line, err = l.solveBlocks(raw_blocks, nil)
	e := s.Close()
	if err == nil {
		target_line = -1
		err = e
	}
	return
}
