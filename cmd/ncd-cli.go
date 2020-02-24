package main

import (
	"github.com/isbm/go-nanoconf"
	daemon "github.com/isbm/uyuni-ncd"
	"github.com/isbm/uyuni-ncd/transport/eventmappers"
	"github.com/urfave/cli/v2"
	"os"
)

func run(ctx *cli.Context) error {
	cfg := nanoconf.NewConfig(ctx.String("config"))
	ncd := daemon.NewNcd()
	ncd.GetTransport().AddNatsServerURL(
		cfg.Find("bus").String("host", ""),
		cfg.Find("bus").DefaultInt("port", "", 4222))

	ncd.GetDBListener().
		SetHost(cfg.Find("db").String("host", "")).
		SetChannel("cluster").
		SetDBName(cfg.Find("db").String("database", "")).
		SetUser(cfg.Find("db").String("user", "")).
		SetPassword(cfg.Find("db").String("password", "")).
		SetSSLMode(false)

	msgmap := eventmappers.NewUyuniEventMapper().
		SetRPCUrl(cfg.Find("api").String("url", "")).
		SetRPCUser(cfg.Find("api").String("user", "")).
		SetRPCPassword(cfg.Find("api").String("password", ""))

	ncd.AddMapper(msgmap).SetLeader(ctx.Bool("leader"))

	ncd.Run()
	return nil
}

func main() {
	appname := "ncd"
	confpath := nanoconf.NewNanoconfFinder(appname).DefaultSetup(nil)

	app := &cli.App{
		Version: "0.1 Alpha",
		Name:    appname,
		Usage:   "Cluster Node Controller Daemon",
		Action:  run,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "config",
				Aliases:  []string{"c"},
				Usage:    "Path to configuration file",
				Required: false,
				Value:    confpath.SetDefaultConfig(confpath.FindFirst()).FindDefault(),
			},
			&cli.BoolFlag{
				Name:    "leader",
				Aliases: []string{"m"},
				Usage:   "Leader mode",
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
	/*
		c := nstcompiler.NewNstCompiler()
		err := c.LoadFile("./test.nst")
		if err != nil {
			panic(err)
		}

		n := nanostate.NewNanostate()
		err = n.Load(c.Tree())
		if err != nil {
			panic(err)
		}

		s := runners.NewSSHRunner().SetSSHHostVerification(false).AddHost("d76.suse.de")
		s.Run(n)
		fmt.Println(s.Response().PrettyJSON())
		fmt.Println("--------------")

		l := runners.NewLocalRunner()
		l.Run(n)
		fmt.Println(l.Response().PrettyJSON())
	*/
}
