package apps

import (
	"encoding/json"

	. "github.com/cloudfoundry/cf-acceptance-tests/cats_suite_helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/workflowhelpers"
	"github.com/cloudfoundry/cf-acceptance-tests/helpers/app_helpers"
	"github.com/cloudfoundry/cf-acceptance-tests/helpers/assets"
	"github.com/cloudfoundry/cf-acceptance-tests/helpers/random_name"
)

var _ = AppsDescribe("Service Discovery", func() {
	var appNameFrontend string
	var appNameBackend string
	var domainName string
	var orgName string
	var spaceName string
	var internalDomainName string
	var internalHostName string
	//LATER: var routeName s
	//tring

	// curlRoute := func(hostName string, path string) string {
	// 	uri := Config.Protocol() + hostName + "." + domainName + path
	// 	curlCmd := helpers.CurlSkipSSL(true, uri).Wait(Config.DefaultTimeoutDuration())
	// 	Expect(curlCmd).To(Exit(0))

	// 	Expect(string(curlCmd.Err.Contents())).To(HaveLen(0))
	// 	return string(curlCmd.Out.Contents())
	// }

	BeforeEach(func() {
		orgName = TestSetup.RegularUserContext().Org
		spaceName = TestSetup.RegularUserContext().Space
		domainName = random_name.CATSRandomName("DOMAIN") + "." + Config.GetAppsDomain()
		workflowhelpers.AsUser(TestSetup.AdminUserContext(), Config.DefaultTimeoutDuration(), func() {
			Expect(cf.Cf("create-shared-domain", domainName).Wait(Config.CfPushTimeoutDuration())).To(Exit(0))
		})

		internalDomainName = "apps.internal"
		internalHostName = "meow"
		appNameFrontend = random_name.CATSRandomName("APP")
		appNameBackend = random_name.CATSRandomName("APP")

		// check internal domain
		sharedDomainBody := cf.Cf("curl", "/v2/shared_domains?q=name:apps.internal").Wait(Config.CfPushTimeoutDuration()).Out.Contents()
		var sharedDomainJSON struct {
			Resources []struct {
				Metadata struct {
					SharedDomainGuid string `json:"guid"`
				} `json:"metadata"`
			} `json:"resources"`
		}
		Expect(json.Unmarshal([]byte(sharedDomainBody), &sharedDomainJSON)).To(Succeed())
		Expect(sharedDomainJSON.Resources[0].Metadata.SharedDomainGuid).ToNot(BeNil())

		//push backend app
		Expect(cf.Cf(
			"push", appNameBackend,
			"--no-start",
			"-b", Config.GetRubyBuildpackName(),
			"-m", DEFAULT_MEMORY_LIMIT,
			"-p", assets.NewAssets().HelloWorld,
			"-d", Config.GetAppsDomain(),
		).Wait(Config.CfPushTimeoutDuration())).To(Exit(0))

		app_helpers.SetBackend(appNameBackend)
		Expect(cf.Cf("start", appNameBackend).Wait(Config.CfPushTimeoutDuration())).To(Exit(0))

		// map internal route to backend app
		Expect(cf.Cf("map-route", appNameBackend, internalDomainName, "--hostname", internalHostName).Wait(Config.CfPushTimeoutDuration())).To(Exit(0))

		// push frontend app
		Expect(cf.Cf(
			"push", appNameFrontend,
			"--no-start",
			"-b", Config.GetBinaryBuildpackName(),
			"-m", DEFAULT_MEMORY_LIMIT,
			"-p", assets.NewAssets().Catnip,
			"-c", "./catnip",
			"-d", Config.GetAppsDomain(),
		).Wait(Config.CfPushTimeoutDuration())).To(Exit(0))

		app_helpers.SetBackend(appNameFrontend)
		Expect(cf.Cf("start", appNameFrontend).Wait(Config.CfPushTimeoutDuration())).To(Exit(0))
	})

	AfterEach(func() {
		app_helpers.AppReport(appNameFrontend, Config.DefaultTimeoutDuration())
		app_helpers.AppReport(appNameBackend, Config.DefaultTimeoutDuration())

		workflowhelpers.AsUser(TestSetup.AdminUserContext(), Config.DefaultTimeoutDuration(), func() {
			Expect(cf.Cf("target", "-o", orgName).Wait(Config.DefaultTimeoutDuration())).To(Exit(0))
			Expect(cf.Cf("delete-shared-domain", domainName, "-f").Wait(Config.DefaultTimeoutDuration())).To(Exit(0))
		})

		Expect(cf.Cf("delete", appNameFrontend, "-f", "-r").Wait(Config.DefaultTimeoutDuration())).To(Exit(0))
		Expect(cf.Cf("delete", appNameBackend, "-f", "-r").Wait(Config.DefaultTimeoutDuration())).To(Exit(0))
	})

	Describe("Adding an internal route on an app", func() {
		FIt("successfully creates a policy", func() {

			workflowhelpers.AsUser(TestSetup.AdminUserContext(), Config.DefaultTimeoutDuration(), func() {
				Expect(cf.Cf("target", "-o", orgName).Wait(Config.DefaultTimeoutDuration())).To(Exit(0))
				Expect(string(cf.Cf("network-policies").Wait(Config.DefaultTimeoutDuration()).Out.Contents())).ToNot(ContainSubstring(appNameBackend))
				Expect(cf.Cf("add-network-policy", appNameFrontend, "--destination-app", appNameBackend).Wait(Config.CfPushTimeoutDuration())).To(Exit(0))
				Expect(string(cf.Cf("network-policies").Wait(Config.DefaultTimeoutDuration()).Out.Contents())).To(ContainSubstring(appNameBackend))
			})
		})
	})
})
