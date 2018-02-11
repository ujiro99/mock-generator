package main

import (
	"os"
	"os/exec"

	log "github.com/Sirupsen/logrus"

	"path/filepath"
	"text/template"

	"fmt"
	"io/ioutil"
	"strings"

	"github.com/ujiro99/mock-generator/ctags"
	"github.com/urfave/cli"
)

type fileType int

const (
	tagFile     = "mock_gen.tag"
	mockPrefix  = "mock_"
	defaultDist = "generated"
	helpTxt     = `
{{.Name}} is c++ mock generator using FakeIt.

Usage:
   {{.HelpName}} [options] [command]
{{if .Commands}}
Commands:
{{range .Commands}}{{if not .HideHelp}}   {{join .Names ", "}}{{ "\t"}}{{.Usage}}{{ "\n" }}{{end}}{{end}}{{end}}{{if .VisibleFlags}}
Options:
   {{range .VisibleFlags}}{{.}}
   {{end}}{{end}}{{if .Copyright}}
Copyright:
   {{.Copyright}}
   {{end}}{{if .Version}}
Version:
   {{.Version}}
   {{end}} {{if len .Authors}}
Author:
   {{range .Authors}}{{ . }}{{end}}
   {{end}}
`

	typeCpp fileType = iota
	typeHpp
)

var (
	cppTmpl *template.Template
	hppTmpl *template.Template
)

type mockParams struct {
	*ctags.File
	MockPath string
	MockDir  string
	FileName string
	Prefix   string
	Type     fileType
}

func main() {
	app := cli.NewApp()
	app.Name = "mock-generator"
	app.Version = "0.0.2"
	app.Author = "Yujiro Takeda"
	app.CustomAppHelpTemplate = helpTxt

	var out string
	var debug bool
	app.Flags = []cli.Flag{
		cli.StringSliceFlag{
			Name:  "exclude, e",
			Usage: "Exclude files and directories matching `PATTERN`.",
		},
		cli.StringFlag{
			Name:        "out, o",
			Usage:       "Output files to `DIRECTORY`.",
			Destination: &out,
		},
		cli.BoolFlag{
			Name:        "debug",
			Usage:       "Show debug logs.",
			Destination: &debug,
		},
	}

	app.Action = func(c *cli.Context) error {
		// initialize
		loadTemplate()
		if out == "" {
			out = defaultDist
		}
		if debug {
			log.SetLevel(log.DebugLevel)
		} else {
			log.SetLevel(log.InfoLevel)
		}

		// exec and parse ctags
		exclude := c.StringSlice("exclude")
		execCtags(exclude)
		bytes, _ := ioutil.ReadFile(tagFile)
		files := ctags.NewParser().Parse(string(bytes))
		os.Remove(tagFile)

		// generate mock files
		fileNum := len(files)
		for i, f := range files {
			fmt.Printf("[%d/%d] generate for %s\n", i+1, fileNum, f.Path)
			generate(*f, out)
		}

		return nil
	}

	app.Run(os.Args)
}

func execCtags(exclude []string) string {
	cmd := []string{"ctags", "-R", "--languages=c,c++", "--extra=+q", "--fields=+KSz", "-f", tagFile}
	if len(exclude) > 0 {
		for _, v := range exclude {
			cmd = append(cmd, "--exclude="+v)
		}
	}
	return output(cmd)
}

func output(cmd []string) string {
	log.Debugln(cmd)
	var c *exec.Cmd
	if len(cmd) >= 2 {
		c = exec.Command(cmd[0], cmd[1:]...)
	} else {
		c = exec.Command(cmd[0])
	}
	// Depending on the environment, fails here.
	s, err := c.Output()
	if err != nil {
		log.Fatal(err)
	}
	return string(s)
}

func loadTemplate() {
	var err error
	cppTmpl, err = template.New("cppTemplate").Parse(cppTemplate)
	if err != nil {
		panic(err)
	}
	hppTmpl, err = template.New("hppTemplate").Parse(hppTemplate)
	if err != nil {
		panic(err)
	}
}

func generate(f ctags.File, out string) {
	removeDuplicatedInclude(&f)
	p := generateParam(f, out)

	err := os.MkdirAll(p.MockDir, 0777)
	if err != nil {
		log.Fatalf("Can't create mock dir: %s\n%s", p.MockDir, err)
	}

	w, err := os.Create(p.MockPath)
	if err != nil {
		log.Fatalf("Can't create mock file: %s\n%s", p.MockPath, err)
	}
	defer w.Close()

	if p.Type == typeCpp {
		err = cppTmpl.Execute(w, p)
	} else {
		err = hppTmpl.Execute(w, p)
	}
	if err != nil {
		log.Fatalf("Can't generate mock: %s\n%s", p.MockPath, err)
	}
}

func generateParam(file ctags.File, out string) mockParams {
	dir, name := filepath.Split(file.Path)
	ext := filepath.Ext(name)

	p := mockParams{File: &file}
	p.MockDir = filepath.Join(out, dir)
	p.MockPath = filepath.Join(out, dir, mockPrefix+name)
	if strings.HasPrefix(ext, ".c") {
		log.Debugf("use type typeCpp for %s", file.Path)
		p.Type = typeCpp
	} else {
		log.Debugf("use type TypeHpp for %s", file.Path)
		p.Type = typeHpp
	}
	p.FileName = name[:len(name)-len(ext)]
	p.Prefix = mockPrefix

	log.WithFields(log.Fields{
		"Classes": len(p.Classes),
		"Funcs":   len(p.Funcs),
	}).Debugln("generated params")
	return p
}

func removeDuplicatedInclude(f *ctags.File) {
	existsInclude := make(map[string]bool, len(f.Classes))
	for _, c := range f.Classes {
		if existsInclude[c.DeclarationFile] {
			c.DeclarationFile = ""
		} else {
			existsInclude[c.DeclarationFile] = true
		}
	}
}
