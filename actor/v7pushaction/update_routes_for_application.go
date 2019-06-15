package v7pushaction

func (actor Actor) UpdateRoutesForApplication(pushPlan PushPlan, eventStream chan<- *PushEvent, progressBar ProgressBar) (PushPlan, Warnings, error) {
	eventStream <- &PushEvent{Plan: pushPlan, Event: CreatingAndMappingRoutes}
	warnings, err := actor.CreateAndMapDefaultApplicationRoute(pushPlan.OrgGUID, pushPlan.SpaceGUID, pushPlan.Application)
	if err != nil {
		return pushPlan, warnings, err
	}
	eventStream <- &PushEvent{Plan: pushPlan, Event: CreatedRoutes}
	return pushPlan, warnings, err
}
