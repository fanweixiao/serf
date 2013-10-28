package command

import (
	"flag"
	"fmt"
	"github.com/hashicorp/serf/cli"
	"strings"
)

// MembersCommand is a Command implementation that queries a running
// Serf agent what members are part of the cluster currently.
type MembersCommand struct{}

func (c *MembersCommand) Help() string {
	helpText := `
Usage: serf members [options]

  Outputs the members of a running Serf agent.

Options:

  -detailed                 Additional information such as protocol verions
                            will be shown.
  -rpc-addr=127.0.0.1:7373  RPC address of the Serf agent.
`
	return strings.TrimSpace(helpText)
}

func (c *MembersCommand) Run(args []string, ui cli.Ui) int {
	var detailed bool
	cmdFlags := flag.NewFlagSet("members", flag.ContinueOnError)
	cmdFlags.Usage = func() { ui.Output(c.Help()) }
	cmdFlags.BoolVar(&detailed, "detailed", false, "detailed output")
	rpcAddr := RPCAddrFlag(cmdFlags)
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	client, err := RPCClient(*rpcAddr)
	if err != nil {
		ui.Error(fmt.Sprintf("Error connecting to Serf agent: %s", err))
		return 1
	}
	defer client.Close()

	members, err := client.Members()
	if err != nil {
		ui.Error(fmt.Sprintf("Error retrieving members: %s", err))
		return 1
	}

	for _, member := range members {
		ui.Output(fmt.Sprintf("%s    %s    %s    %s",
			member.Name, member.Addr, member.Status, member.Role))

		if detailed {
			ui.Output(fmt.Sprintf("    Protocol Version: %d",
				member.DelegateCur))
			ui.Output(fmt.Sprintf("    Available Protocol Range: [%d, %d]",
				member.DelegateMin, member.DelegateMax))
		}
	}

	return 0
}

func (c *MembersCommand) Synopsis() string {
	return "Lists the members of a Serf cluster"
}
