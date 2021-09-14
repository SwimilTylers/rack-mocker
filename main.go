package rack_mocker

import (
	"flag"
	"go.etcd.io/etcd/pkg/fileutil"
	"log"
	"os"
)

var bindSocket = flag.String("bind-socket", "rmocker.sock", "specify the socket file mocker binds to")
var port = flag.Int("port", 39253, "specify the port file mocker listens to")
var rootDirectory = flag.String("root-directory", "rmocker-root", "specify the root directory managed by mocker")

func main() {
	flag.Parse()

	
	// fs := os.DirFS(*rootDirectory)
}
