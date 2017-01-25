package main

////////////////////////////////////////////////////
/// Command line interface for interacting with ///
//              blacklist database             //
///////////////////////////////////////////////
import (
	"os"

	"github.com/ocmdev/rita-blacklist"
	"github.com/urfave/cli"
)

var (
	// hostFlag specifies a database host
	hostFlag = cli.StringFlag{
		Name:  "host",
		Usage: "Database `HOST`",
		Value: "localhost",
	}

	// portFlag specifies a database port
	portFlag = cli.IntFlag{
		Name:  "port,p",
		Usage: "Database `PORT`",
		Value: 0,
	}

	targetFlag = cli.StringFlag{
		Name:  "target,t",
		Usage: "Lookup the given `TARGET` in the blacklist database",
		Value: "",
	}

	allCommands = []cli.Command{
		cli.Command{
			Name:  "lookup",
			Usage: "lookup host in blacklist database",
			Flags: []cli.Flag{
				hostFlag,
				portFlag,
				targetFlag,
			},
			Action: func(c *cli.Context) error {
				bl := blacklist.NewBlackList()
				bl.Init(c.String("host"), c.Int("port"), "blacklistedHosts")
				hosts := []string{c.String("target")}
				bl.CheckHosts(hosts, "blacklistedHosts")
				return nil
			},
		},
	}
)

func main() {
	app := cli.NewApp()
	app.Name = "blacklist"
	app.Usage = "Simple blacklist lookup cli"

	app.Version = "0.0.1"

	app.Commands = allCommands

	app.Run(os.Args)
}
