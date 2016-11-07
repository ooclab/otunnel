package cmd

import "github.com/urfave/cli"

// Commands 表示子命令集合
var Commands = []cli.Command{
	{
		Name:   "listen",
		Usage:  "Listen as a server, wait connects from clients.",
		Flags:  listenFlags,
		Action: listenAction,
	},
	{
		Name:   "connect",
		Usage:  "connect to a server",
		Flags:  connectFlags,
		Action: connectAction,
	},
	// {
	// 	Name:  "client",
	// 	Usage: "client tool for adjust port binding, ...",
	// },
	// {
	// 	Name:  "forward, nat",
	// 	Usage: "端口转发（实现 iptables NAT 功能）",
	// },
	// {
	// 	Name:  "list",
	// 	Usage: "show all tunnels",
	// },
	// {
	// 	// 参考： iptables-save
	// 	Name:  "save",
	// 	Usage: "dump current config to stdin",
	// },
	// {
	// 	// 参考： iptables-restore
	// 	Name:  "restore",
	// 	Usage: "read config from stdin",
	// },
	// {
	// 	// 参考： iptables-restore
	// 	Name:   "genca",
	// 	Usage:  "generate ca.pem and ca.key",
	// 	Flags:  gencaFlags,
	// 	Action: gencaAction,
	// },
	// {
	// 	// 参考： iptables-restore
	// 	Name:   "example",
	// 	Usage:  "show usage example",
	// 	Action: exampleAction,
	// },
}
