//Package gocmd provides simple abstraction for the go command line tool.
package gocmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

// reflect some basic go operations

//Wrapper around some go operations
type GoEnv struct {
	gopath string
}

func NewGoEnv(gopath string) *GoEnv {
	return &GoEnv{
		gopath: gopath,
	}
}

//BuildEnv gets the current os.Environ() map and override the keys define in vals
func BuildEnv(vals map[string]string) []string {
	current := os.Environ()
	newenv := make([]string, 0, len(current))
	for _, v := range current {
		parts := strings.SplitN(v, "=", 2)
		k := parts[0]
		if val, ok := vals[k]; ok { // overwrite it
			newenv = append(newenv, fmt.Sprintf("%s=%s", k, val))
			vals[k] = "" // mark it has deleted
		} else {
			newenv = append(newenv, fmt.Sprintf("%s=%s", k, parts[1]))
		}
	}
	// now append the "new" ones
	for k, val := range vals {
		if len(val) != 0 {
			newenv = append(newenv, fmt.Sprintf("%s=%s", k, val))
		}

	}

	return newenv
}

//Join joins a path like variable with some addition
func Join(path string, elements ...string) string {
	if len(path) == 0 {
		return strings.Join(elements, string(os.PathListSeparator))
	} else {
		return path + string(os.PathListSeparator) + strings.Join(elements, string(os.PathListSeparator))
	}
	panic("unreachable statement")

}

//Install wrap the go install command. 
// For the moment only -a option is available
func (g *GoEnv) Install(root string, all bool, ldflags string) (err error) {
	var cmd *exec.Cmd
	//ldflags = "'" + ldflags + "'"
	if all {
		cmd = exec.Command("go", "install", "-a", "-ldflags", ldflags, "./src/...")
	} else {
		//fmt.Printf("ldflags=%s", ldflags)
		cmd = exec.Command("go", "install", "-ldflags", ldflags, "./src/...")
	}

	// extend the current env with my GOPATH variable
	locals := map[string]string{
		"GOPATH": Join(root, g.gopath),
	}

	cmd.Env = BuildEnv(locals)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = root // asbolute path of the project
	err = cmd.Run()
	return
}

//InstallTest wrap the go test -c command to compile test exe for a given package 
// For the moment only -a option is available
func (g *GoEnv) InstallTest(root, pack string) (err error) {
	var cmd *exec.Cmd
	cmd = exec.Command("go", "test", "-c", pack)
	// extend the current env with my GOPATH variable
	locals := map[string]string{
		"GOPATH": Join(root, g.gopath),
	}

	cmd.Env = BuildEnv(locals)
	cmd.Stdin = os.Stdin
	cmd.Stdout = ioutil.Discard //Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = root // asbolute path of the project
	err = cmd.Run()
	if err != nil {
		return
	}
	return
}

// Wrapper around go test command
func (g *GoEnv) Test(root, wd string, args []string) (err error) {

	arguments := make([]string, 0, len(args)+2)
	arguments = append(arguments, "test")
	arguments = append(arguments, args...)

	cmd := exec.Command("go", arguments...) //"./src/...")

	locals := map[string]string{
		"GOPATH": Join(g.gopath, root),
	}

	cmd.Env = BuildEnv(locals)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = wd // absolute path of the project
	err = cmd.Run()
	return
}

//Wrapper around go get command
func (g *GoEnv) Get(pack string) {

	cmd := exec.Command("go", "get", pack)

	locals := map[string]string{
		"GOPATH": g.gopath,
	}
	//fmt.Printf("GOPATH = %v\n", g.gopath)
	cmd.Env = BuildEnv(locals)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = g.gopath // asbolute path of the project
	if err := cmd.Run(); err != nil {
		fmt.Printf("%v\n", err)
	}
}
