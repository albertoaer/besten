package modules

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	. "github.com/besten/internal/lexer"
	"github.com/besten/internal/parser"
	. "github.com/besten/internal/runtime"
)

func prettyTokens(blocks []Block, tabs string) {
	for _, b := range blocks {
		fmt.Printf(tabs+"%v\n", b.Tokens)
		prettyTokens(b.Children, tabs+"  ")
	}
}

type fileSource struct {
	path string
}

func (f *fileSource) Origin() string {
	return f.path
}

func (f *fileSource) GetSource() (io.ReadCloser, error) {
	return os.Open(f.path)
}

type storedModule struct {
	identifier int
	available  bool
	requesters []int
	path       string
	queue      chan interface{}
	exclusion  sync.Mutex
	parser     *parser.Parser
	result     error
}

type Modules struct {
	modules  []*storedModule
	entries  map[string]int
	modulemx sync.Mutex
	symbols  map[string]*Symbol
	symbolmx sync.Mutex
}

func New() *Modules {
	return &Modules{make([]*storedModule, 0), make(map[string]int),
		sync.Mutex{}, make(map[string]*Symbol), sync.Mutex{}}
}

func (m *Modules) NewId() int {
	var id int
	id = len(m.modules)
	m.modules = append(m.modules, nil)
	return id
}

func existsFile(path string) bool {
	var err error
	var abspath string
	if abspath, err = filepath.Abs(path); err != nil {
		return false
	}
	_, err = os.Stat(abspath)
	return err == nil || !errors.Is(err, os.ErrNotExist)
}

func (m *Modules) LoadModule(requester int, path string) (parser.Module, error) {
	var abspath string
	{
		m.modulemx.Lock()
		m.modules[requester].exclusion.Lock()
		folder := filepath.Dir(m.modules[requester].path)
		abspath = filepath.Join(folder, path)
		m.modules[requester].exclusion.Unlock()
		m.modulemx.Unlock()
	}
	if !existsFile(abspath) {
		return nil, fmt.Errorf("Module %s does not exists", path)
	}
	var fdata fs.FileInfo
	var err error
	if fdata, err = os.Stat(abspath); err != nil {
		return nil, err
	}
	if !fdata.IsDir() {
		p, e := m.FileParser(requester, abspath)
		if e != nil {
			return nil, e
		}
		return p.GetModule(), nil
	} else if !existsFile(filepath.Join(abspath, "mod.bst")) {
		return nil, errors.New("Expecting mod.bst file for folder module")
	}
	var files []fs.FileInfo
	if files, err = ioutil.ReadDir(path); files != nil {
		return nil, err
	}
	for _, file := range files {
		if file.IsDir() {
			go m.LoadModule(requester, filepath.Join(path, file.Name()))
		} else if file.Name() != "mod.bst" {
			go m.FileParser(requester, filepath.Join(path, file.Name()))
		}
	}
	p, e := m.FileParser(requester, filepath.Join(abspath, "mod.bst"))
	if e != nil {
		return nil, e
	}
	return p.GetModule(), nil
}

func (m *Modules) checkModuleExistence(requester int, path string, md **storedModule) bool {
	m.modulemx.Lock()
	defer m.modulemx.Unlock()
	idx, existence := m.entries[path]
	if !existence {
		idx = m.NewId()
		m.entries[path] = idx
		m.modules[idx] = &storedModule{idx, false, []int{requester}, path, make(chan interface{}), sync.Mutex{}, nil, nil}
	}
	*md = m.modules[idx]
	return existence
}

func (m *Modules) isRequesting(target, requester int) bool {
	m.modulemx.Lock()
	defer m.modulemx.Unlock()
	m.modules[target].exclusion.Lock()
	defer m.modules[target].exclusion.Unlock()
	for i := range m.modules[target].requesters {
		if m.modules[target].requesters[i] == requester {
			return true
		}
	}
	return false
}

func (m *Modules) FileParser(requester int, path string) (*parser.Parser, error) {
	var err error
	if path, err = filepath.Abs(path); err != nil {
		return nil, err
	}
	if filepath.Ext(path) != ".bst" {
		return nil, errors.New("File does not have besten extension")
	}
	var md *storedModule
	if m.checkModuleExistence(requester, path, &md) {
		md.exclusion.Lock()
		md.requesters = append(md.requesters, requester)
		md.exclusion.Unlock()
		if m.isRequesting(requester, md.identifier) {
			return nil, errors.New("Circular dependency")
		}
		var parser *parser.Parser
		var err error
		md.exclusion.Lock()
		if md.available {
			parser = md.parser
			err = md.result
		}
		md.exclusion.Unlock()
		if err != nil || parser == nil {
			return parser, err
		}
		<-md.queue
		md.exclusion.Lock()
		if !md.available {
			err = errors.New("File was no treaty")
		} else {
			parser = md.parser
			err = md.result
		}
		md.exclusion.Unlock()
		return parser, err
	}
	filesrc := fileSource{path}
	blocks, err := LexerFor(&filesrc).GetBlocks()
	var module_parser *parser.Parser
	if err == nil {
		module_parser = parser.NewParser(path, md.identifier, m)
		err = module_parser.ParseCode(blocks)
	}
	md.exclusion.Lock()
	md.available = true
	md.parser = module_parser
	md.result = err
	close(md.queue)
	md.exclusion.Unlock()
	return module_parser, err
}

func (m *Modules) MainFile(name string) (symbols map[string]Symbol, cname string, err error) {
	module_parser, e := m.FileParser(-1, name)
	if e != nil {
		err = e
		return
	}
	cname, err = module_parser.GetSymbolNameFor("main", false, []parser.OBJType{parser.VecOf(parser.Str)})
	if err != nil {
		return
	}
	symbols = m.collectSymbols()
	return
}
