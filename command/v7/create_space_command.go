package v7

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/clock"
)

//go:generate counterfeiter . CreateSpaceActor

type CreateSpaceActor interface {
	CreateSpace(spaceName, orgGUID string) (v7action.Space, v7action.Warnings, error)
	GetOrganizationByName(orgName string) (v7action.Organization, v7action.Warnings, error)
}

type CreateSpaceCommand struct {
	RequiredArgs    flag.Space  `positional-args:"yes"`
	Organization    string      `short:"o" description:"Organization"`
	usage           interface{} `usage:"CF_NAME create-space SPACE [-o ORG]"`
	relatedCommands interface{} `related_commands:"set-space-isolation-segment, space-quotas, spaces, target"`

	UI          command.UI
	Config      command.Config
	Actor       CreateSpaceActor
	SharedActor command.SharedActor
}

func (cmd *CreateSpaceCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	sharedActor := sharedaction.NewActor(config)
	cmd.SharedActor = sharedActor

	ccClient, uaaClient, err := shared.NewClients(config, ui, true, "")
	if err != nil {
		return err
	}
	cmd.Actor = v7action.NewActor(ccClient, config, sharedActor, uaaClient, clock.NewClock())
	return nil
}

func (cmd CreateSpaceCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	var (
		orgName, orgGUID string
	)

	if cmd.Organization == "" {
		_, err = cmd.SharedActor.RequireTargetedOrg()
		if err != nil {
			return err
		}
		orgName = cmd.Config.TargetedOrganization().Name
		orgGUID = cmd.Config.TargetedOrganization().GUID
	} else {
		orgName = cmd.Organization
		org, warnings, err := cmd.Actor.GetOrganizationByName(orgName)
		cmd.UI.DisplayWarnings(warnings)
		if err != nil {
			return err
		}
		orgGUID = org.GUID
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	spaceName := cmd.RequiredArgs.Space

	cmd.UI.DisplayTextWithFlavor("Creating space {{.Space}} in org {{.Organization}} as {{.User}}...",
		map[string]interface{}{
			"User":         user.Name,
			"Space":        spaceName,
			"Organization": orgName,
		})

	_, warnings, err := cmd.Actor.CreateSpace(spaceName, orgGUID)

	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		if _, ok := err.(actionerror.SpaceAlreadyExistsError); ok {
			cmd.UI.DisplayText(err.Error())
			cmd.UI.DisplayOK()
			return nil
		}
		return err
	}

	cmd.UI.DisplayOK()

	cmd.UI.DisplayText(`TIP: Use 'cf target -o "{{.Organization}}" -s "{{.Space}}"' to target new space`,
		map[string]interface{}{
			"Organization": orgName,
			"Space":        spaceName,
		})

	return nil
}
