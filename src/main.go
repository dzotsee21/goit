package main

import (
	"goit/src/api"
	"log"
	"os"
	"regexp"
)

type CommandFunc func(args []interface{}) error

var goit = map[string]CommandFunc{
	"init": func(args []interface{}) error {
		api.Init(args[0])
		return nil
	},
	"add": func(args []interface{}) error {
		api.Add(args[0].(string))
		return nil
	},
	"rm": func(args []interface{}) error {
		api.Rm(args[0].(string))
		return nil
	},
	"commit": func(args []interface{}) error {
		api.Commit(map[string]string{"m": args[0].(string)})
		return nil
	},
	"branch": func(args []interface{}) error {
		api.Branch(args[0])
		return nil
	},
	"checkout": func(args []interface{}) error {
		api.Checkout(args[0].(string))
		return nil
	},
	"diff": func(args []interface{}) error {
		api.Diff(args[0], args[1])
		return nil
	},
	"remote": func(args []interface{}) error {
		api.Remote(args[0].(string), args[1].(string), args[2].(string))
		return nil
	},
	"fetch": func(args []interface{}) error {
		api.Fetch(args[0], args[1])
		return nil
	},
	"merge": func(args []interface{}) error {
		api.Merge(args[0].(string))
		return nil
	},
	"pull": func(args []interface{}) error {
		api.Pull(args[0].(string), args[1].(string))
		return nil
	},
	"push": func(args []interface{}) error {
		api.Push(args[0], args[1], args[2].(map[string]string))
		return nil
	},
	"status": func(args []interface{}) error {
		api.Status()
		return nil
	},
	"clone": func(args []interface{}) error {
		api.Clone(args[0].(string), args[1].(string), args[2])
		return nil
	},
}

func parseOptions(argv []string) map[string]interface{} {
	var name string

	opts := make(map[string]interface{})
	for _, arg := range argv {
		re := regexp.MustCompile("^-/")
		if re.Match([]byte(arg)) {
			re := regexp.MustCompile("^-+/,")
			name = re.ReplaceAllString(arg, "")
			opts[name] = true
		}
		if name != "" {
			opts[name] = arg
		} else {
			_, exists := opts["_"]
			if !exists {
				opts["_"] = []interface{}{}
			}
			opts["_"] = append(opts["_"].([]interface{}), arg)
		}

	}

	if opts != nil {
		return opts
	} else {
		return map[string]interface{}{"_": []string{}}
	}
}

func runCli(argv []string) {
	opts := parseOptions(argv)
	commandName := opts["_"].([]interface{})[2]

	if commandName == nil {
		log.Fatal("you must specify a Goit cmd to run")
	} else {
		re := regexp.MustCompile("-")
		commandFnName := re.ReplaceAllString(commandName.(string), "_")
		fn := goit[commandFnName]

		if fn == nil {
			log.Fatal("'" + commandFnName + "' is not a Goit cmd")
		} else {
			commandArgs := opts["_"].([]interface{})[3:]

			err := fn(commandArgs)
			if err != nil {
				log.Fatal(err)
			}
			return
		}
	}

}

func main() {
	runCli(os.Args)
}
