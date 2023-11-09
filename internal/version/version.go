package version

import "fmt"

var version string
var commit string

func Get() string {
	return fmt.Sprintf("%s-%s", version, commit)
}
