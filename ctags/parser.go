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
	p.methodPattern = regexp.MustCompile(`^(.+)\t(.+)\t/\^(.+)\(.*\$/;"\tkind:(\S+)\tclass:(\S+)\tsignature:\((.*)\)`)
	p.funcPattern = regexp.MustCompile(`^(.+)\t(.+)\t/\^(.+)\(.*\$/;"\tkind:(\S+)\tsignature:\((.*)\)`)
	p.linePattern = regexp.MustCompile(`^(\S+)\s+\S+`)
	return &p
}

// ctagsParser implements Parser
type ctagsParser struct {
	// regex patterns for analyze ctags.
	classPattern  *regexp.Regexp
	methodPattern *regexp.Regexp
	funcPattern   *regexp.Regexp
	linePattern   *regexp.Regexp
}

// Parse a lines of ctags.
func (p *ctagsParser) Parse(tags string) []*File {

	files := make([]*File, 0)
	classes := make([]*Class, 0)

	lines := strings.Split(tags, "\n")
	for _, line := range lines {
		if strings.Index(line, typeClass) >= 0 {

			m := p.classPattern.FindStringSubmatch(line)
			if len(m) > keyNumClass {

				log.WithFields(log.Fields{
					name:     m[1],
					filePath: m[2],
					raw:      m[3],
					kind:     m[4],
				}).Debugln("class")

				f := p.findFile(&files, m[2])
				if f == nil {
					f = &File{
						Path:    m[2],
						Classes: make([]*Class, 0),
						Funcs:   make([]Func, 0),
					}
					files = append(files, f)
				}
				c := p.findClass(&classes, m[1])
				if c == nil {
					c = &Class{
						Name:  m[1],
						Funcs: make([]Func, 0),
					}
					classes = append(classes, c)
					_, c.DeclarationFile = filepath.Split(m[2])
				}
				f.Classes = append(f.Classes, c)
			}

		} else if strings.Index(line, typeFunc) >= 0 {
			m := p.methodPattern.FindStringSubmatch(line)
			if len(m) > keyNumMethod {

				log.WithFields(log.Fields{
					name:      m[1],
					filePath:  m[2],
					raw:       m[3],
					kind:      m[4],
					className: m[5],
					signature: m[6],
				}).Debugln("func")

				f := p.findFile(&files, m[2])
				if f == nil {
					f = &File{
						Path:    m[2],
						Classes: make([]*Class, 0),
						Funcs:   make([]Func, 0),
					}
					files = append(files, f)
				}

				c := p.findClass(&classes, m[5])
				if c == nil {
					c = &Class{
						Name:  m[5],
						Funcs: make([]Func, 0),
					}
					classes = append(classes, c)
				}
				found := false
				for _, c := range f.Classes {
					found = found || (c.Name == m[5])
				}
				if !found {
					f.Classes = append(f.Classes, c)
				}

				method := strings.Split(m[1], "::")
				if len(method) >= 2 {
					log.Debugf("find method %s", m[1])

					fn := p.findFunc(&c.Funcs, method[1])
					if fn == nil {
						ml := p.linePattern.FindStringSubmatch(m[3])
						if len(ml) >= 2 {
							c.Funcs = append(c.Funcs, Func{
								Name:         method[1],
								Signature:    ml[0],
								Return:       ml[1],
								Args:         p.extractArgs(m[6]),
								ArgWithTypes: m[6],
							})

							log.WithFields(log.Fields{
								"Signature":    ml[0],
								"Return":       ml[1],
								"Args":         p.extractArgs(m[6]),
								"ArgWithTypes": m[6],
							}).Debugln("class raw")
						}
					}
				} else {
					fn := p.findFunc(&c.Funcs, m[1])
					if fn != nil {
						continue
					}
					log.Debugf("ignore constructor %s(%s)", m[1], m[6])

					// c.Funcs = append(c.Funcs, Func{
					// 	Name:         m[1],
					// 	Signature:    m[6],
					// 	Return:       "",
					// 	Args:         p.extractArgs(m[6]),
					// 	ArgWithTypes: m[6],
					// })

					// log.WithFields(log.Fields{
					// 	"Signature":    m[6],
					// 	"Return":       "",
					// 	"Args":         p.extractArgs(m[6]),
					// 	"ArgWithTypes": m[6],
					// }).Debugln("constructor ")

				}
			} else {
				m := p.funcPattern.FindStringSubmatch(line)
				if len(m) > keyNumFunc {
					log.WithFields(log.Fields{
						name:      m[1],
						filePath:  m[2],
						raw:       m[3],
						kind:      m[4],
						signature: m[5],
					}).Debugln("func")

					f := p.findFile(&files, m[2])
					if f == nil {
						f = &File{
							Path:    m[2],
							Classes: make([]*Class, 0),
							Funcs:   make([]Func, 0),
						}
						files = append(files, f)
					}
					fn := p.findFunc(&f.Funcs, m[1])
					if fn == nil {
						ml := p.linePattern.FindStringSubmatch(m[3])
						if len(ml) >= 2 {
							f.Funcs = append(f.Funcs, Func{
								Name:         m[1],
								Signature:    ml[0],
								Return:       ml[1],
								Args:         p.extractArgs(m[5]),
								ArgWithTypes: m[5],
							})
							log.WithFields(log.Fields{
								"Signature":    ml[0],
								"Return":       ml[1],
								"Args":         p.extractArgs(m[5]),
								"ArgWithTypes": m[5],
							}).Debugln("file raw")
						}
					}
				}
			}
		}
	}
	return files
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
