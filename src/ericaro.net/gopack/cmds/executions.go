package cmds

//Executions commands

import (
	. "ericaro.net/gopack"
	"ericaro.net/gopack/gocmd"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

func init() {
	Reg(
		&Compile,
		&Test,
		//add package-compile, compile a package in the local repo
	)

}

// The flag package provides a default help printer via -h switch
var compileAllFlag *bool
var compileOfflineFlag *bool
var compileUpdateFlag *bool
var compileSkipTestFlag *bool
var compileLDFlag *string
var Compile = Command{
	Name:      `compile`,
	Alias:     `c`,
	Category:  CompileCategory,
	UsageLine: ``,
	Short:     `Compile project`,
	Long: `Computes current project dependencies as a GOPATH variable (accessible through the p Option),
       and then run go install on the project.`,
	RequireProject: true,
	FlagInit: func(Compile *Command) {
		compileAllFlag = Compile.Flag.Bool("a", false, "all. Force rebuilding of packages that are already up-to-date.")
		compileOfflineFlag = Compile.Flag.Bool("o", false, "offline. Do not use the network to look for missing dependencies.")
		compileUpdateFlag = Compile.Flag.Bool("u", false, "update. Look for updated version of dependencies.")
		compileLDFlag = Compile.Flag.String("ldflags", "", "pass ldflags to all sub commands")
		compileSkipTestFlag = Compile.Flag.Bool("s", false, "skip. Skip Test compilation")
	},
	Run: func(Compile *Command) (err error) {
		// parse dependencies, and build the gopath
		dependencies, err := Compile.Repository.ResolveDependencies(Compile.Project, *compileOfflineFlag, *compileUpdateFlag)
		if err != nil {
			ErrorStyle.Printf("Error Resolving project's dependencies:\n    \u21b3 %v", err)
			return
		}
		// run the go build command for local src, and with the appropriate gopath
		gopath, err := Compile.Repository.GoPath(dependencies)

		goEnv := gocmd.NewGoEnv(gopath)
		err = goEnv.Install(Compile.Project.WorkingDir(), *compileAllFlag, *compileLDFlag)

		if !*compileSkipTestFlag {

			//compute the GOOS that will be used by the test compiler (mainly to be crossplatform compliant
			goos := runtime.GOOS //default value
			if os.Getenv("GOOS") != "" {
				goos = os.Getenv("GOOS")
			}

			goarch := runtime.GOARCH //default value
			if os.Getenv("GOARCH") != "" {
				goarch = os.Getenv("GOARCH")
			}

			log.Printf("Parsing packages to test-compile\n")
			packages := Compile.Project.Packages() // list all packages

			root := Compile.Project.WorkingDir()
			for _, p := range packages {
				log.Printf("compiling %s tests\n", p)
				err = goEnv.InstallTest(root, p)
				if err != nil {
					return err
				}
				//infer the exe name
				name := filepath.Base(p) + ".test"
				if goos == "windows" {
					name += ".exe"
				}
				// move the exe to the appropriate place
				dst := filepath.Join(root, "bin", goos+"_"+goarch, name)
				src := filepath.Join(root, name)
				if FileExists(src) {
					os.MkdirAll(filepath.Dir(dst), os.ModeDir|os.ModePerm)
					err = os.Rename(src, dst)
					//_, err = CopyFile(dst, src)
					if err != nil {
						return
					}
				}
			}
		}

		// also provide a go run equivalent 
		return
	},
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////:

var testWatchFlag *time.Duration
var Test = Command{
	Name:           `test`,
	Alias:          `t`,
	Category:       CompileCategory,
	UsageLine:      ``,
	Short:          `Run go test`,
	Long:           `Run go test on the whole project.`, // TODO add options to select the package to be executed
	RequireProject: true,
	FlagInit: func(Test *Command) {
		testWatchFlag = Test.Flag.Duration("w", 0, "watch. Repeat the command for ever every watched seconds")
	},
	Run: func(Test *Command) (err error) {

		// parse dependencies, and build the gopath
		dependencies, err := Test.Repository.ResolveDependencies(Test.Project, true, false)
		if err != nil {
			ErrorStyle.Printf("Error Resolving project's dependencies:\n    \u21b3 %v", err)
			return
		}

		// run the go build command for local src, and with the appropriate gopath
		gopath, err := Test.Repository.GoPath(dependencies)
		if err != nil {
			ErrorStyle.Printf("Invalid dependency:\n    \u21b3 %v", err)
			return
		}
		// make two cases, either I'm on the root dir, then the package is ./src/... or I'm within the project 
		// and the packages path is ./...

		wd, err := os.Getwd()
		// use cwd as root, unless it is in the workding dir
		var wdi, pwdi os.FileInfo
		wdi, err = os.Stat(wd)
		pwdi, err = os.Stat(Test.Project.WorkingDir())

		goEnv := gocmd.NewGoEnv(gopath)
		var args []string
		if os.SameFile(wdi, pwdi) { // if I'm on the project root dir, sources are in ./src/...
			args = []string{"./src/..."}
		} else { // otherwise assume that you know what you are doing, and just recuse in ./...
			args = []string{"./..."}
		}

		// check for -watch option
		if *testWatchFlag > 0 {
			c := time.Tick(100 * time.Millisecond)
			next := time.Now()
			for now := range c {
				if now.After(next) {
					next = now.Add(*testWatchFlag)
					NormalStyle.Clear()
					err = goEnv.Test(Test.Project.WorkingDir(), wd, args)
					fmt.Printf("\n")
				}
				fmt.Printf("■")
			}

		} else {
			// single run
			err = goEnv.Test(Test.Project.WorkingDir(), wd, args)
		}

		return
	},
}
