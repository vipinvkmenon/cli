package v7pushaction_test

import (
	. "code.cloudfoundry.org/cli/actor/v7pushaction"
	"code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Actor", func() {
	var (
		actor *Actor
		plan  PushPlan
	)

	BeforeEach(func() {
		actor, _, _ = getTestPushActor()
	})

	Describe("PreparePushPlanSequence", func() {
		It("is a list of functions for preparing the push plan", func() {
			Expect(actor.PreparePushPlanSequence).To(matchers.MatchFuncsByName(
				SetupApplicationForPushPlan,
				SetupDockerImageCredentialsForPushPlan,
				SetupBitsPathForPushPlan,
				SetupDropletPathForPushPlan,
				actor.SetupAllResourcesForPushPlan,
				SetupDeploymentStrategyForPushPlan,
				SetupNoStartForPushPlan,
				SetupNoWaitForPushPlan,
				SetupSkipRouteCreationForPushPlan,
				SetupScaleWebProcessForPushPlan,
				SetupUpdateWebProcessForPushPlan,
			))
		})
	})

	Describe("ChangeApplicationSequence", func() {
		BeforeEach(func() {
			plan = PushPlan{
				ApplicationNeedsUpdate:            true,
				SkipRouteCreation:                 true,
				DockerImageCredentialsNeedsUpdate: false,
			}
		})

		It("returns a sequence including the required functions from all three sequences", func() {
			Expect(actor.ChangeApplicationSequence(plan)).To(matchers.MatchFuncsByName(
				actor.UpdateApplication,
				actor.CreateBitsPackageForApplication,
				actor.StagePackageForApplication,
				actor.SetDropletForApplication,
				actor.RestartApplication,
			))
		})
	})
})
