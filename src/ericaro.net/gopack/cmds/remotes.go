package cmds

import (
	"bytes"
	. "ericaro.net/gopack"
	"ericaro.net/gopack/oauth"
	"ericaro.net/gopack/protocol"
	"ericaro.net/gopack/semver"
	"fmt"
	"log"
	"net/url"
)

func init() {
	Reg(
		&Serve,
		&Push,
		&AddRemote,
		&RemoveRemote,
	)

}

var serverAddrFlag *string

var Serve = Command{
	Name:      `serve`,
	Alias:     `serve`,
	UsageLine: `ADDR`,
	Short:     `Serve local repository as an http server`,
	Long: `Serve local repository as an http remote repository
       so that others can get latest updates, or push new releases.
       ADDR usually ':8080' `,
	RequireProject: false, // false if we add the options to set which the local repo
	FlagInit: func(Serve *Command) {
		serverAddrFlag = Serve.Flag.String("s", ":8080", "Serve the current local repository as a remote one for others to use.")
	},
	Run: func(Serve *Command) (err error) {

		// run the go build command for local src, and with the appropriate gopath

		server := HttpServer{
			Local: *Serve.Repository,
		}
		fmt.Printf("starting server %s\n", *serverAddrFlag)
		server.Start(*serverAddrFlag)
		return

	},
}

//var deployAddrFlag *string = Push.Flag.String("to", "central", "deploy to a specific remote repository.")
//var pushRecursiveFlag *bool = Push.Flag.Bool("r", false, "Also pushes package's dependencies.")
var pushExecutables *bool
var Push = Command{
	Name:      `push`,
	Alias:     `push`,
	UsageLine: `REMOTE PACKAGE VERSION`,
	Short:     `Push a project in a remote repository`,
	Long: `Push a project in a remote repository
       REMOTE  a remote name in the remote list
       PACKAGE a package available in the local repository (use search to list them)
       VERSION a semantic version of the PACKAGE available in the local repository
       
       If binaries are pushed too, then they will be available at:
       
       <remote-url>/get?n=<package-name>&v=<package-version>&goos=<os>&goarch=<architecture>&exe=<executable-name>

       Note, that the executable name is the full name, meaning with a trailing .exe if the goos=windows"
       The list of available executable is also available at:
       
       <remote-url>/list?n=<package-name>&v=<package-version>&goos=<os>&goarch=<architecture>
       
       the server returns a list, in json format of download url.
       
       
       
       `,
	RequireProject: false,
	FlagInit: func(Push *Command) {
		pushExecutables = Push.Flag.Bool("x", false, "pushes executables too.")
	},
	Run: func(Push *Command) (err error) {
		rem := Push.Flag.Arg(0)
		remote, err := Push.Repository.Remote(rem)
		if err != nil {
			ErrorStyle.Printf("Unknown Remote %s.\n    \u21b3 %s\n", rem, err)

			fmt.Printf("Available remotes are:\n")
			for _, r := range Push.Repository.Remotes() {
				u := r.Path()
				fmt.Printf("    %-40s %s\n", r.Name(), u.String())
			}
			return
		}

		version, err := semver.ParseVersion(Push.Flag.Arg(2))
		if err != nil {
			ErrorStyle.Printf("Invalid Version \"%s\".\n    \u21b3 %s\n", Push.Flag.Arg(2), err)
			return
		}

		// now look for the real package in the local repo
		pkg, err := Push.Repository.FindPackage(*NewProjectID(Push.Flag.Arg(1), version))
		if err != nil {
			ErrorStyle.Printf("Cannot find Package %s %s in Local Repository %s.\n    \u21b3 %s\n", Push.Flag.Arg(1), Push.Flag.Arg(2), Push.Repository.Root(), err)
			// TODO as soon as I've got some search capability display similar results
			return
		}

		// build its ID 
		tm := pkg.Timestamp()
		pid := protocol.PID{
			Name:      Push.Flag.Arg(1),
			Version:   version,
			Token:     remote.Token(),
			Timestamp: &tm,
		}

		// read it in memory (tar.gz)

		buf := new(bytes.Buffer)
		pkg.Pack(buf) // pack either exec or src
		// and finally push the buffer
		log.Printf("pushing sources\n")
		err = remote.Push(pid, buf) // either exec or src

		if *pushExecutables {
			log.Printf("pushing executables")
			buf := new(bytes.Buffer)
			pkg.PackExecutables(buf) // pack both exec or src

			// and finally push the buffer
			err = remote.PushExecutables(pid, buf) // either exec or src		
		}

		if err != nil {
			ErrorStyle.Printf("Error from the remote while pushing.\n    \u21b3 %s\n", err)
			// TODO as soon as I've got some search capability display similar results
			return
		}
		SuccessStyle.Printf("Success\n")
		return
	},
}

//var Get = Command{
//	Name:           `goget`,
//	Alias:          `gg`,
//	UsageLine:      `<goget package>`,
//	Short:          `Run go get a package and install it`,
//	Long:           `Run go get a package and install it`,
//	Run:            func(Get *Command) { panic("not yet implemented") },
//	RequireProject: false,
//}

////////////////////////////////////////////////////////////////////////////////////////

var oauthFlag *bool     // OAuth flag value
var base64Token *string // Simple token value

var AddRemote = Command{
	Name:      `radd`,
	Alias:     `r+`,
	Category:  RemoteCategory,
	UsageLine: `NAME URL`,
	Short:     `Add a remote server.`,
	Long: `Remote server can be used to publish or receive other's code.
       NAME    local alias for this remote
       URL     full URL to the remote server.
               file:// and http:// are actually supported. For http, see 'gpk serve'`,
	RequireProject: false,
	FlagInit: func(AddRemote *Command) {
		oauthFlag = AddRemote.Flag.Bool("o", false, "OAuth, when the remote must be accessed using OAuth 1.0 authentification.")
		base64Token = AddRemote.Flag.String("b", "", "Base64 token, when the remote must be accessed using a base64 token")
	},
	Run: func(AddRemote *Command) (err error) {

		// Check arguments:
		if len(AddRemote.Flag.Args()) != 2 {
			ErrorStyle.Printf("Illegal arguments count\n")
			return
		}

		if len(*base64Token) > 0 && *oauthFlag {
			ErrorStyle.Printf("Illegal arguments combinaison, -o and -b options can't be used together.\n")
			return
		}

		// Retrieve NAME & URL values:
		name, remote := AddRemote.Flag.Arg(0), AddRemote.Flag.Arg(1)

		var token *protocol.Token // nil by default
		// Handling Base64 token:
		if len(*base64Token) > 0 {
			// When -b flag:
			token, err = protocol.ParseStdToken(*base64Token)
			if err != nil {
				ErrorStyle.Printf("Invalid token syntax, please enter a valid token base64- RFC 4648 Encoded array of bytes.\n")
				return
			}
		}

		// Handling OAuth token:
		if *oauthFlag {
			// When -o option, the method RequestOAuthToken() start a procedure to request
			// an OAuth token:
			token, err = oauth.RequestOAuthToken(name, remote)
			if err != nil {
				ErrorStyle.Printf("Failed to request OAuth token.\n    \u21b3 %s\n", err)
				return
			}

			// Hack to force the creation of an OAuth client:
			remote = "oauth:" + remote
		}

		// Verifying that the remote is valid URL:
		u, err := url.Parse(remote)
		if err != nil {
			ErrorStyle.Printf("Invalid URL passed as a remote Repository.\n    \u21b3 %s\n", err)
			return
		}
		if u.String() == "" {
			ErrorStyle.Printf("Invalid URL passed as a remote Repository.\n    \u21b3 %s\n", remote)
			return
		}

		// Instantiate a remote client:
		client, err := protocol.NewClient(name, *u, token)
		if err != nil {
			ErrorStyle.Printf("Failed to create the a client for this remote:\n    \u21b3 %s\n", err)
			return
		}
		if client == nil {
			ErrorStyle.Printf("Failed to create the a client for this remote:\n    \u21b3 %s\n", remote)
			return
		}

		// Update configuration:
		err = AddRemote.Repository.RemoteAdd(client)
		if err != nil {
			ErrorStyle.Printf("%s\n", err)
			return
		}
		AddRemote.Repository.Write()

		// Display a success trace
		stoken := ""
		if token != nil {
			stoken = fmt.Sprintf("%s", token)
		}
		SuccessStyle.Printf("       +%s %s %s\n", name, u, stoken)
		return
	},
}

var RemoveRemote = Command{
	Name:           `rremove`,
	Alias:          `r-`,
	Category:       RemoteCategory,
	UsageLine:      `NAME`,
	Short:          `Remove a Remote`,
	Long:           ``,
	RequireProject: false,
	Run: func(RemoveRemote *Command) (err error) {

		if len(RemoveRemote.Flag.Args()) != 1 {
			RemoveRemote.Flag.Usage()
			return
		}

		name := RemoveRemote.Flag.Arg(0)
		ref, err := RemoveRemote.Repository.RemoteRemove(name)
		if err != nil {
			ErrorStyle.Printf("Cannot Remove remote %s\n    \u21b3 %s", name, err)
			return
		}
		if ref != nil {
			u := ref.Path()
			SuccessStyle.Printf("Removed Remote %s %s\n", ref.Name(), u.String())
			RemoveRemote.Repository.Write()
		} else {
			ErrorStyle.Printf("Nothing to Remove\n")
		}
		RemoveRemote.Repository.Write()
		return
	},
}
