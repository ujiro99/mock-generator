package ctags

import (
	"path/filepath"
	"regexp"
	"strings"

	log "github.com/Sirupsen/logrus"
)

const (
	// ctags fields
	name      = "name"
	filePath  = "filePath"
	raw       = "raw"
	kind      = "kind"
	className = "className"
	signature = "signature"

	// tag types
	typeClass = "kind:class"
	typeFunc  = "kind:function"

	inClass = "class:"
)

// key order for analyze ctags.
var (
	keyOrderClass  = []string{name, filePath, raw, kind}
	keyNumClass    = len(keyOrderClass)
	keyOrderMethod = []string{name, filePath, raw, kind, className, signature}
	keyNumMethod   = len(keyOrderMethod)
	keyOrderFunc   = []string{name, filePath, raw, kind, signature}
	keyNumFunc     = len(keyOrderFunc)
)

// File represents a cpp/hpp file.
type File struct {
	Path    string
	Classes []*Class
	Funcs   []Func
}

// Class represents a class.
type Class struct {
	Name            string
	Funcs           []Func
	DeclarationFile string
}

// Func represents a function.
type Func struct {
	Name         string
	Signature    string
	Return       string
	Args         string
	ArgWithTypes string
}

// Parser parse ctags to struct.
type Parser interface {
	Parse(tags string) []*File
}

// NewParser creates Parser instance.
func NewParser() Parser {
	p := ctagsParser{}
	p.classPattern = regexp.MustCompile(`^(.+)\t(.+)\t/\^(.+).*\$/;"\tkind:(\S+)`)
	p.methodPattern = regexp.MustCompile(`^.+::(\S+)\t(.+)\t/\^(.+)\(.*\$/;"\tkind:(\S+)\tclass:(\S+)\tsignature:\((.*)\)`)
	p.funcPattern = regexp.MustCompile(`^(.+)\t(.+)\t/\^(.+)\(.*\$/;"\tkind:(\S+)\tsignature:\((.*)\)`)
	p.rawPattern = regexp.MustCompile(`^(\S+)\s+\S+`)
	return &p
}

// ctagsParser implements Parser
type ctagsParser struct {
	// regex patterns for analyze ctags.
	classPattern  *regexp.Regexp
	methodPattern *regexp.Regexp
	funcPattern   *regexp.Regexp
	rawPattern    *regexp.Regexp
}

// Parse a lines of ctags.
func (p *ctagsParser) Parse(tags string) []*File {
	files := make([]*File, 0)
	classes := make([]*Class, 0)
	lines := strings.Split(tags, "\n")

	for _, line := range lines {
		if strings.Index(line, typeClass) >= 0 {
			p.parseClass(line, &files, &classes)
		} else if strings.Index(line, typeFunc) >= 0 {
			p.parseFunction(line, &files, &classes)
		}
	}

	return files
}

func (p *ctagsParser) parseClass(line string, files *[]*File, classes *[]*Class) {
	m := p.classPattern.FindStringSubmatch(line)
	if len(m) > keyNumClass {
		log.WithFields(p.toMap(m, keyOrderClass)).Debugln("class")
		f := p.ensureFile(files, m[2])
		c := p.ensureClass(classes, m[1], m[2])
		f.Classes = append(f.Classes, c)
	}
}

func (p *ctagsParser) parseFunction(line string, files *[]*File, classes *[]*Class) {
	if strings.Index(line, inClass) > 0 {
		m := p.methodPattern.FindStringSubmatch(line)
		if len(m) > keyNumMethod {
			log.WithFields(p.toMap(m, keyOrderMethod)).Debugln("func")
			f := p.ensureFile(files, m[2])
			c := p.ensureClass(classes, m[5], m[2])
			if p.findClass(&f.Classes, m[5]) == nil {
				f.Classes = append(f.Classes, c)
			}
			p.parseMethod(m[3], m[1], m[6], &c.Funcs)
		}
	} else {
		m := p.funcPattern.FindStringSubmatch(line)
		if len(m) > keyNumFunc {
			log.WithFields(p.toMap(m, keyOrderFunc)).Debugln("func")
			f := p.ensureFile(files, m[2])
			p.parseMethod(m[3], m[1], m[5], &f.Funcs)
		}
	}
}

func (p *ctagsParser) parseMethod(raw string, name string, signature string, funcs *[]Func) {
	fn := p.findFunc(funcs, name)
	if fn != nil {
		return
	}

	m := p.rawPattern.FindStringSubmatch(raw)
	if len(m) >= 2 {
		log.WithFields(log.Fields{
			"Signature":    m[0],
			"Return":       m[1],
			"Args":         p.extractArgs(signature),
			"ArgWithTypes": signature,
		}).Debugln("class raw")

		*funcs = append(*funcs, Func{
			Name:         name,
			Signature:    m[0],
			Return:       m[1],
			Args:         p.extractArgs(signature),
			ArgWithTypes: signature,
		})
	}
}

func (p *ctagsParser) ensureClass(classes *[]*Class, name string, filePath string) *Class {
	c := p.findClass(classes, name)
	if c == nil {
		c = &Class{
			Name:  name,
			Funcs: make([]Func, 0),
		}
		*classes = append(*classes, c)
		_, c.DeclarationFile = filepath.Split(filePath)
	}
	return c
}

func (p *ctagsParser) ensureFile(files *[]*File, filePath string) *File {
	f := p.findFile(files, filePath)
	if f == nil {
		f = &File{
			Path:    filePath,
			Classes: make([]*Class, 0),
			Funcs:   make([]Func, 0),
		}
		*files = append(*files, f)
	}
	return f
}

func (p *ctagsParser) findFile(ls *[]*File, key string) *File {
	for _, v := range *ls {
		if v.Path == key {
			return v
		}
	}
	return nil
}

func (p *ctagsParser) findClass(ls *[]*Class, key string) *Class {
	for _, v := range *ls {
		if v.Name == key {
			return v
		}
	}
	return nil
}

func (p *ctagsParser) findFunc(ls *[]Func, key string) *Func {
	for _, v := range *ls {
		if v.Name == key {
			return &v
		}
	}
	return nil
}

func (p *ctagsParser) extractArgs(arguments string) string {
	args := strings.Split(arguments, ",")
	vars := make([]string, len(args))
	for i, a := range args {
		if t := strings.Split(a, " "); len(t) >= 2 {
			a = t[len(t)-1]
		}
		if t := strings.Split(a, "*"); len(t) >= 2 {
			a = strings.Trim(t[len(t)-1], " ")
		}
		vars[i] = strings.TrimRight(a, "[]")
	}
	return strings.Join(vars, ", ")
}

func (p *ctagsParser) toMap(arry []string, keys []string) map[string]interface{} {
	m := make(map[string]interface{}, len(keys))
	for i, k := range keys {
		m[k] = arry[i+1]
	}
	return m
}
