package version

import (
	"fmt"
	"net/http"
	// "net/http"

	"github.com/spf13/cobra"
)

var (
	// VersionCmd prints out the current sdk version
	VersionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the app version",
		Run:   printVersion,
	}
)

func GetVersion() string {
	vForbole := VersionF
	vCosmos := VersionC
	v := "Forbole version: " + vForbole + "\n" + "Cosmos version: " + vCosmos
	if GitCommitForbole != "" || GitCommitCosmos != "" {
		v = "Forbole version: " + vForbole + "-" + GitCommitForbole + "\n" + "Cosmos version: " + vCosmos + "-" + GitCommitCosmos
	}
	return v
}

// CMD
func printVersion(cmd *cobra.Command, args []string) {
	v := GetVersion()
	fmt.Println(v)
}

// version REST handler endpoint
func RequestHandler(w http.ResponseWriter, r *http.Request) {
	v := GetVersion()
	w.Write([]byte(v))
}

// var (
// 	// VersionCmd prints out the current sdk version
// 	VersionCmd = &cobra.Command{
// 		Use:   "version",
// 		Short: "Print the app version",
// 		Run:   printVersion,
// 	}
// )

// // return version of CLI/node and commit hash
// func GetVersion() string {
// 	v := Version
// 	if GitCommit != "" {
// 		v = v + "-" + GitCommit
// 	}
// 	return v
// }

// // CMD
// func printVersion(cmd *cobra.Command, args []string) {
// 	v := GetVersion()
// 	fmt.Println(v)
// }
