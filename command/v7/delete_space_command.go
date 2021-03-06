package v7

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/clock"
)

//go:generate counterfeiter . DeleteSpaceActor

type DeleteSpaceActor interface {
	DeleteSpaceByNameAndOrganizationName(spaceName string, orgName string) (v7action.Warnings, error)
}

type DeleteSpaceCommand struct {
	RequiredArgs flag.Space  `positional-args:"yes"`
	Force        bool        `short:"f" description:"Force deletion without confirmation"`
	Org          string      `short:"o" description:"Delete space within specified org"`
	usage        interface{} `usage:"CF_NAME delete-space SPACE [-o ORG] [-f]"`

	Config      command.Config
	UI          command.UI
	SharedActor command.SharedActor
	Actor       DeleteSpaceActor
}

func (cmd *DeleteSpaceCommand) Setup(config command.Config, ui command.UI) error {
	cmd.Config = config
	cmd.UI = ui
	sharedActor := sharedaction.NewActor(config)
	cmd.SharedActor = sharedActor

	ccClient, uaaClient, err := shared.NewClients(config, ui, true, "")
	if err != nil {
		return err
	}
	cmd.Actor = v7action.NewActor(ccClient, config, sharedActor, uaaClient, clock.NewClock())

	return nil
}

func (cmd DeleteSpaceCommand) Execute(args []string) error {
	var (
		err     error
		orgName string
	)

	if cmd.Org == "" {
		err = cmd.SharedActor.CheckTarget(true, false)
		orgName = cmd.Config.TargetedOrganization().Name
	} else {
		err = cmd.SharedActor.CheckTarget(false, false)
		orgName = cmd.Org
	}

	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	if !cmd.Force {
		cmd.UI.DisplayText("This action impacts all resources scoped to this space, including apps, service instances, and space-scoped service brokers.")
		promptMessage := "Really delete the space {{.SpaceName}}?"
		deleteSpace, promptErr := cmd.UI.DisplayBoolPrompt(false, promptMessage, map[string]interface{}{"SpaceName": cmd.RequiredArgs.Space})

		if promptErr != nil {
			return promptErr
		}

		if !deleteSpace {
			cmd.UI.DisplayText("'{{.TargetSpace}}' has not been deleted.",
				map[string]interface{}{
					"TargetSpace": cmd.RequiredArgs.Space,
				})
			return nil
		}
	}

	cmd.UI.DisplayTextWithFlavor("Deleting space {{.TargetSpace}} in org {{.TargetOrg}} as {{.CurrentUser}}...",
		map[string]interface{}{
			"TargetSpace": cmd.RequiredArgs.Space,
			"TargetOrg":   orgName,
			"CurrentUser": user.Name,
		})

	warnings, err := cmd.Actor.DeleteSpaceByNameAndOrganizationName(cmd.RequiredArgs.Space, orgName)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()

	if cmd.Config.TargetedOrganization().Name == orgName &&
		cmd.Config.TargetedSpace().Name == cmd.RequiredArgs.Space {
		cmd.Config.UnsetSpaceInformation()
		cmd.UI.DisplayText("TIP: No space targeted, use '{{.CfTargetCommand}}' to target a space.",
			map[string]interface{}{"CfTargetCommand": cmd.Config.BinaryName() + " target -s"})
	}

	return nil
}
