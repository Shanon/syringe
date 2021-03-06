package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/tanksuzuki/syringe/backend"
	"github.com/tanksuzuki/syringe/log"
	"github.com/tanksuzuki/syringe/template"
	"golang.org/x/crypto/ssh/terminal"
)

const (
	VERSION string = "0.1.2"
)

type flagSet struct {
	Backend    string            `short:"b" long:"backend" description:"Backend type" default:"toml" default-mask:"toml" env:"SY_BACKEND"`
	Template   string            `short:"t" long:"template" description:"Template File"`
	VarFile    string            `short:"V" long:"variable-file" description:"Variable File" default:"-"`
	Debug      bool              `long:"debug" description:"Enable debug logging" env:"SY_DEBUG"`
	DelimLeft  string            `long:"delim-left" env:"SY_DELIML" description:"Template start delimiter" default:"{{" default-mask:"{{"`
	DelimRight string            `long:"delim-right" env:"SY_DELIMR" description:"Template end delimiter" default:"}}" default-mask:"}}"`
	Help       bool              `short:"h" long:"help" description:"Show this help"`
	Variable   map[string]string `short:"v" long:"variable" description:"Set key/values (format key:value)"`
	Version    bool              `long:"version" description:"Show version information"`
}

type cli struct {
	inStream             io.Reader
	outStream, errStream io.Writer
}

func main() {
	c := &cli{inStream: os.Stdin, outStream: os.Stdout, errStream: os.Stderr}
	os.Exit(c.run(os.Args[1:]))
}

func (c *cli) run(args []string) int {
	log.SetLevel("info")

	flag, args, err := parseArgs(args)
	if err != nil {
		log.Error(fmt.Sprintf("%s", err), c.errStream)
		return 1
	}

	if flag.Debug {
		log.SetLevel("debug")
	}

	log.Debug(fmt.Sprintf("Parsed flags: %+v", flag), c.outStream)
	log.Debug(fmt.Sprintf("Parsed args: %+v", args), c.outStream)

	switch {
	case flag.Help:
		fmt.Fprintln(c.outStream, help())
		return 1
	case flag.Version:
		fmt.Fprintln(c.outStream, version())
		return 1
	case len(args) == 0:
		log.Error("syringe requires a minimum of 1 argument. Please see 'syringe --help'.", c.errStream)
		return 1
	}

	tmplFile := args[0]
	if flag.Template != "" {
		tmplFile = flag.Template
	}

	keyValue, err := c.loadKeyValue( tmplFile, flag.VarFile, flag.Variable, flag.Backend, args )
	if err != nil {
		log.Error( fmt.Sprintf("%s", err ), c.errStream )
		return 1
	}

	log.Debug("Merge the key/values and template...", c.outStream)

	bufTmpl, err := c.loadTemplate( tmplFile )
	if err != nil {
		log.Error(fmt.Sprintf("%s", err), c.errStream)
		return 1
	}

	merged, err := template.Merge(string(bufTmpl), flag.DelimLeft, flag.DelimRight, keyValue)
	if err != nil {
		log.Error(fmt.Sprintf("%s", err), c.errStream)
		return 1
	}

	log.Debug("Output the merged text", c.outStream)
	fmt.Fprintf(c.outStream, "%s", merged)

	return 0
}

func ( c *cli) loadKeyValue( tmplFile string, varFile string, variable map[string]string, bType string, args []string ) ( map[string]interface{}, error ) {
	keyValue := map[string]interface{}{}
	
	var kvBuf []byte
	if varFile == "-" {
		if tmplFile != "-"  && !terminal.IsTerminal(0) && fmt.Sprintf("%s", c.inStream) != "" {
			log.Debug("Get the key/values from pipe...", c.outStream)
			pipe, err := ioutil.ReadAll(c.inStream)
			if err != nil {
				return keyValue, err
			}
			kvBuf = pipe
		}
	} else {
		log.Debug( "Get the key.values from file...", c.outStream)
		fbuf, err := ioutil.ReadFile( varFile )
		if err != nil {
			return keyValue, err
		}
		kvBuf = fbuf
	}
	bufKeyValue, err := backend.GetKeyValueFromString(bType, string(kvBuf))
	if err != nil {
		return keyValue, err
	}
	keyValue = mergeMapIfaceIface(keyValue, bufKeyValue)
	log.Debug(fmt.Sprintf("Current key/value %+v", keyValue), c.outStream)

	log.Debug("Get the key/value from backends...", c.outStream)
	backendKeyValue, err := backend.GetKeyValueFromBackends(bType, args[1:])
	if err != nil {
		return keyValue, err
	}
	keyValue = mergeMapIfaceIface(keyValue, backendKeyValue)
	log.Debug(fmt.Sprintf("Current key/value %+v", keyValue), c.outStream)

	if len(variable) > 0 {
		log.Debug("Get the key/values from v flag...", c.outStream)
		mergeMapIfaceString(keyValue, variable)
		log.Debug(fmt.Sprintf("Current key/values %+v", keyValue), c.outStream)
	}
	
	return keyValue, nil
}

func ( c *cli ) loadTemplate( tmplFile string ) ( []byte, error ) {
	var bufTmpl []byte
	if tmplFile == "-" && !terminal.IsTerminal(0) && fmt.Sprintf("%s", c.inStream) != "" {
		log.Debug( "Get the template from pipe...", c.outStream )
		pipe, err := ioutil.ReadAll( c.inStream )
		if err != nil {
			return pipe, err
		}
		bufTmpl = pipe
	} else {
		log.Debug( "Get the template from file...", c.outStream )
		buf, err := ioutil.ReadFile( tmplFile )
		if err != nil {
			return buf, err
		}
		bufTmpl = buf
	}
	return bufTmpl, nil
}

func mergeMapIfaceIface(m1, m2 map[string]interface{}) map[string]interface{} {
	for k, _ := range m2 {
		m1[k] = m2[k]
	}
	return m1
}

func mergeMapIfaceString(m1 map[string]interface{}, m2 map[string]string) map[string]interface{} {
	for k, _ := range m2 {
		m1[k] = m2[k]
	}
	return m1
}

func parseArgs(args []string) (flagSet, []string, error) {
	var flag flagSet
	p := flags.NewParser(&flag, flags.PassDoubleDash)
	args, err := p.ParseArgs(args)
	return flag, args, err
}

func help() string {
	var out io.Writer = new(bytes.Buffer)
	var f flagSet
	p := flags.NewParser(&f, flags.PassDoubleDash)
	p.Name = "syringe"
	p.Usage = "[options] <template> [<backend>...]"
	p.WriteHelp(out)
	return fmt.Sprintf("%s", out)
}

func version() string {
	return fmt.Sprintf("syringe version %s", VERSION)
}
