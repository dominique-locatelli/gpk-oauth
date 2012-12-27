package semver

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const (
	digits = `(\d+)?(?:\.(\d+)(?:\.(\d+))?)?`
	sub    = `[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*`
	all    = `%s(?:\-?(%s))?(?:\+(%s))?`
)

var (
	SemVer, _ = regexp.Compile(fmt.Sprintf(all, digits, sub, sub))
)

//Version is a struct that hold all [http://semver.org/ semantic version] components.
type Version struct {
	major, minor, patch uint32
	pre, build          string
}

// TODO add tests and methods to this struct
// TODO add persistence format control ( string back and forth is a good objective)

func (v Version) String() (version string) {
	version = fmt.Sprintf("%d.%d.%d", v.major, v.minor, v.patch)
	if v.major == 0 && v.minor == 0 && v.patch == 0 {
		version = fmt.Sprintf("%s", v.pre)
	} else {
		if len(v.pre) != 0 {
			version += fmt.Sprintf("-%s", v.pre)
		}
	}
	if len(v.build) != 0 {
		version += fmt.Sprintf("+%s", v.build)
	}
	return
}
func (v Version) Digits() (major, minor, patch uint32) {
	return v.major, v.minor, v.patch
}
func (v Version) PreRelease() string {
	return v.pre
}
func (v Version) Build() string {
	return v.build
}

func (v Version) IsSnapshot() bool {
	return v.major == 0 && v.minor == 0 && v.patch == 0
}

func atoi(s string) uint32 {
	i, _ := strconv.ParseUint(s, 10, 8)
	return uint32(i)
}

func ParseVersion(v string) (version Version, err error) {
	v = strings.Trim(v, " {}[]\"'")
	parts := SemVer.FindStringSubmatch(v)
	//fmt.Printf("%24s -> %v\n",v, parts[1:])
	version = Version{
		major: atoi(parts[1]),
		minor: atoi(parts[2]),
		patch: atoi(parts[3]),
		pre:   parts[4],
		build: parts[6],
	}
	return
}