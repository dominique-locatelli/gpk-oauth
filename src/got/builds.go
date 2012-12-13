package main

import (
	"fmt"
	"got.ericaro.net/got"
	"strings"
	"os"
)

func init() {
	Reg(
		&Compile,
		&Install,
		&Get,
	)

}

var Compile = Command{
	Name:           `compile`,
	UsageLine:      ``,
	Short:          `Compile the current project`,
	Long:           `Computes current project dependencies as a GOPATH variable (accessible through the p Option), and then compile the project`,
	call:           func(c *Command) { c.Compile() },
	RequireProject: true,
}


var Install = Command{
	Name:           `install`,
	UsageLine:      `<version>`,
	Short:          `Install the current project in the local repository`,
	Long:           `Install the current project in the local repository`,
	call:           func(c *Command) { c.Install() },
	RequireProject: true,
}


var Get = Command{
	Name:           `get`,
	UsageLine:      `<goget package>`,
	Short:          `go get a package and install it`,
	Long:           `go get a package and install it`,
	call:           func(c *Command) { c.Get() },
	RequireProject: true,
}

// The flag package provides a default help printer via -h switch
var compileVersionFlag *bool = Compile.Flag.Bool("v", false, "Print the version number.")
var compileReleaseFlag *bool = Compile.Flag.Bool("r", false, "Build using only release dependencies.")
var compileOfflineFlag *bool = Compile.Flag.Bool("o", false, "Try to find missing dependencies at http://got.ericaro.net")
var compileUpdateFlag *bool = Compile.Flag.Bool("u", false, "Look for updated version of dependencies")
var compilePathOnlyFlag *bool = Compile.Flag.Bool("p", false, "Does not run the compile, just print the gopath (suitable for scripting for instance: alias GP='export GOPATH=`got compile -p`' )")

func (c *Command) Compile() {

	// parse dependencies, and build the gopath
	dependencies, err := c.Repository.FindProjectDependencies(c.Project, !*compileReleaseFlag, *compileOfflineFlag, *compileUpdateFlag)
	if err != nil {
		fmt.Printf("Error Parsing the project's dependencies: %v", err)
		return
	}

	// run the go build command for local src, and with the appropriate gopath
	gopath, err := c.Repository.GoPath(dependencies)
	if *compilePathOnlyFlag {
		path := make([]string, 0,2)
		if gopath != "" {
			path = append(path, gopath)
		}
		path = append(path, c.Project.Root)
		prjgopath := strings.Join(path, string(os.PathListSeparator))

		fmt.Print(prjgopath)
		return
	} else {
		goEnv := got.NewGoEnv(gopath)
		goEnv.Install(c.Project.Root)
	}

}


var installReleaseFlag *bool = Install.Flag.Bool("r", false, "Install as a Release in the local Repository")

func (c *Command) Install() {
	version := got.ParseVersion(c.Flag.Arg(0) )
	c.Repository.InstallProject(c.Project, version ,!*installReleaseFlag)
}



func (c *Command) Get() { 
	for _,p:= range c.Flag.Args() {
		
		// make a package group and name (based on package name)
		// then assign a 0.0.0.0 version and snapshot
		c.Repository.GoGetInstall(p)
		// TO BE CONTINUED
		
	}
}



