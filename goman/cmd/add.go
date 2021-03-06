package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"unicode"

	"github.com/govenue/goman"
)

func init() {
	addCmd.Flags().StringVarP(&packageName, "package", "t", "", "target package name (e.g. github.com/geego/gean)")
	addCmd.Flags().StringVarP(&parentName, "parent", "p", "rootCmd", "variable name of parent command for this command")
}

var packageName, parentName string

var addCmd = &goman.Command{
	Use:     "add [command name]",
	Aliases: []string{"command"},
	Short:   "Add a command to a goman Application",
	Long: `Add (goman add) will create a new command, with a license and
the appropriate structure for a goman-based CLI application,
and register it to its parent (default rootCmd).

If you want your command to be public, pass in the command name
with an initial uppercase letter.

Example: goman add server -> resulting in a new cmd/server.go`,

	Run: func(cmd *goman.Command, args []string) {
		if len(args) < 1 {
			er("add needs a name for the command")
		}

		var project *Project
		if packageName != "" {
			project = NewProject(packageName)
		} else {
			wd, err := os.Getwd()
			if err != nil {
				er(err)
			}
			project = NewProjectFromPath(wd)
		}

		cmdName := validateCmdName(args[0])
		cmdPath := filepath.Join(project.CmdPath(), cmdName+".go")
		createCmdFile(project.License(), cmdPath, cmdName)

		fmt.Fprintln(cmd.OutOrStdout(), cmdName, "created at", cmdPath)
	},
}

// validateCmdName returns source without any dashes and underscore.
// If there will be dash or underscore, next letter will be uppered.
// It supports only ASCII (1-byte character) strings.
func validateCmdName(source string) string {
	i := 0
	l := len(source)
	// The output is initialized on demand, then first dash or underscore
	// occurs.
	var output string

	for i < l {
		if source[i] == '-' || source[i] == '_' {
			if output == "" {
				output = source[:i]
			}

			// If it's last rune and it's dash or underscore,
			// don't add it output and break the loop.
			if i == l-1 {
				break
			}

			// If next character is dash or underscore,
			// just skip the current character.
			if source[i+1] == '-' || source[i+1] == '_' {
				i++
				continue
			}

			// If the current character is dash or underscore,
			// upper next letter and add to output.
			output += string(unicode.ToUpper(rune(source[i+1])))
			// We know, what source[i] is dash or underscore and source[i+1] is
			// uppered character, so make i = i+2.
			i += 2
			continue
		}

		// If the current character isn't dash or underscore,
		// just add it.
		if output != "" {
			output += string(source[i])
		}
		i++
	}

	if output == "" {
		return source // source is initially valid name.
	}
	return output
}

func createCmdFile(license License, path, cmdName string) {
	template := `{{comment .copyright}}
{{if .license}}{{comment .license}}{{end}}

package {{.cmdPackage}}

import (
	"fmt"

	"github.com/govenue/goman"
)

// {{.cmdName}}Cmd represents the {{.cmdName}} command
var {{.cmdName}}Cmd = &goman.Command{
	Use:   "{{.cmdName}}",
	Short: "A brief description of your command",
	Long: ` + "`" + `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

goman is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a goman application.` + "`" + `,
	Run: func(cmd *goman.Command, args []string) {
		fmt.Println("{{.cmdName}} called")
	},
}

func init() {
	{{.parentName}}.AddCommand({{.cmdName}}Cmd)

	// Here you will define your flags and configuration settings.

	// goman supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// {{.cmdName}}Cmd.PersistentFlags().String("foo", "", "A help for foo")

	// goman supports local flags which will only run when this command
	// is called directly, e.g.:
	// {{.cmdName}}Cmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
`

	data := make(map[string]interface{})
	data["copyright"] = copyrightLine()
	data["license"] = license.Header
	data["cmdPackage"] = filepath.Base(filepath.Dir(path)) // last dir of path
	data["parentName"] = parentName
	data["cmdName"] = cmdName

	cmdScript, err := executeTemplate(template, data)
	if err != nil {
		er(err)
	}
	err = writeStringToFile(path, cmdScript)
	if err != nil {
		er(err)
	}
}
