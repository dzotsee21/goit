package main

import (
	"fmt"
	"goit/src/api"
	"log"
	"os"
	"regexp"
)

type CommandFunc func(args []interface{}) error

var goit = map[string]CommandFunc{
	"init": func(args []interface{}) error {
		if len(args) == 1 {
			api.Init(args[0])
		} else {
			fmt.Println("usage: goit init <bare: false|true>")
		}

		return nil
	},
	"add": func(args []interface{}) error {
		if len(args) == 1 {
			api.Add(args[0].(string))
		} else {
			fmt.Println("usage: goit add <files|folders|.>")
		}

		return nil
	},
	"rm": func(args []interface{}) error {
		if len(args) == 1 {
			api.Rm(args[0].(string))
		} else {
			fmt.Println("usage: goit rm <file|folders|.>")
		}

		return nil
	},
	"commit": func(args []interface{}) error {
		if len(args) == 1 {
			val := api.Commit(map[string]string{"m": args[0].(string)})
			fmt.Println(val)
		} else {
			fmt.Println("usage: goit commit <commit_msg>")
		}

		return nil
	},
	"branch": func(args []interface{}) error {
		if len(args) == 1 {
			val := api.Branch(args[0])
			fmt.Println(val)
		} else {
			fmt.Println("usage: goit branch <name>")
		}

		return nil
	},
	"checkout": func(args []interface{}) error {
		if len(args) == 1 {
			val := api.Checkout(args[0].(string))
			fmt.Println(val)			
		} else {
			fmt.Println("usage: goit checkout <branch_name>")
		}

		return nil
	},
	"diff": func(args []interface{}) error {
		if len(args) == 2 {
			val := api.Diff(args[0], args[1])
			fmt.Println(val)
		} else {
			fmt.Println("usage: goit diff <ref1> <ref2>")
		}

		return nil
	},
	"remote": func(args []interface{}) error {
		if len(args) == 3 {
			val := api.Remote(args[0].(string), args[1].(string), args[2].(string))
			fmt.Println(val)
		} else {
			fmt.Println("usage: goit remote <cmd> <name> <remote_path>")
		}

		return nil
	},
	"fetch": func(args []interface{}) error {
		if len(args) == 2 {
			val := api.Fetch(args[0], args[1])
			fmt.Println(val)			
		}

		return nil
	},
	"merge": func(args []interface{}) error {
		if len(args) == 1 {
			val := api.Merge(args[0].(string))
			fmt.Println(val)
		} else {
			fmt.Println("usage: goit merge <ref>")
		}

		return nil
	},
	"pull": func(args []interface{}) error {
		if len(args) == 2 {
			val := api.Pull(args[0].(string), args[1].(string))
			fmt.Println(val)
		} else {
			fmt.Println("usage: goit pull <remote> <branch_name>")
		}

		return nil
	},
	"push": func(args []interface{}) error {
		if len(args) == 3 {
			val := api.Push(args[0], args[1], args[2].(string))
			fmt.Println(val)
		} else {
			fmt.Println("usage: goit push <remote> <branch_name> <cmd>")
		}

		return nil
	},
	"status": func(args []interface{}) error {
		if len(args) == 0 {
			val := api.Status()
			fmt.Println(val)			
		}

		return nil
	},
	"clone": func(args []interface{}) error {
		if len(args) == 3 {
			api.Clone(args[0].(string), args[1].(string), args[2])
		} else {
			fmt.Println("usage: goit clone <remote_path> <target_path> <bare: false|true>")
		}

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

	if len(opts["_"].([]interface{})) <= 1 {
		green := "\033[32m"
		reset := "\033[0m"

		fmt.Println(green + `
		██████╗  ██████╗ ██╗████████╗
		██╔════╝ ██╔═══██╗██║╚══██╔══╝
		██║  ███╗██║   ██║██║   ██║   
		██║   ██║██║   ██║██║   ██║   
		╚██████╔╝╚██████╔╝██║   ██║   
		╚═════╝  ╚═════╝ ╚═╝   ╚═╝   
		` + reset)
	} else {
		commandName := opts["_"].([]interface{})[1]

		if commandName == nil {
			log.Fatal("you must specify a Goit cmd to run")
		} else {
			re := regexp.MustCompile("-")
			commandFnName := re.ReplaceAllString(commandName.(string), "_")
			fn := goit[commandFnName]

			if fn == nil {
				log.Fatal("'" + commandFnName + "' is not a Goit cmd")
			} else {
				commandArgs := opts["_"].([]interface{})[2:]

				err := fn(commandArgs)
				if err != nil {
					log.Fatal(err)
				}
				return
			}
		}
	}

}

func main() {
	runCli(os.Args)
}
