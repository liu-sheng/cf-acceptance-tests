package windows

import (
	"fmt"
	"strings"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
	. "github.com/cloudfoundry/cf-acceptance-tests/cats_suite_helpers"
	"github.com/cloudfoundry/cf-acceptance-tests/helpers/app_helpers"
	"github.com/cloudfoundry/cf-acceptance-tests/helpers/assets"
	"github.com/cloudfoundry/cf-acceptance-tests/helpers/random_name"
	"github.com/cloudfoundry/cf-acceptance-tests/helpers/skip_messages"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

var _ = WindowsDescribe("Context Paths", func() {
	var (
		appName1 string

		appName2 string
		app2Path = "/app2"
		appName3 string
		app3Path = "/app3/long/sub/path"
		hostname string
	)

	BeforeEach(func() {
		if !Config.GetUseWindowsContextPath() {
			Skip(skip_messages.SkipWindowsContextPathsMessage)
		}
		domain := Config.GetAppsDomain()

		appName1 = random_name.CATSRandomName("APP")
		Expect(cf.Cf("push",
			appName1,
			"-s", Config.GetWindowsStack(),
			"-b", Config.GetHwcBuildpackName(),
			"-m", DEFAULT_MEMORY_LIMIT,
			"-p", assets.NewAssets().Nora,
			"-d", Config.GetAppsDomain()).Wait(Config.CfPushTimeoutDuration())).To(Exit(0))

		// .NET Apps must be pushed with --no-start --no-route before running cf map-route
		appName2 = random_name.CATSRandomName("APP")
		Expect(cf.Cf("push",
			appName2,
			"--no-start",
			"--no-route",
			"-s", Config.GetWindowsStack(),
			"-b", Config.GetHwcBuildpackName(),
			"-m", DEFAULT_MEMORY_LIMIT,
			"-p", assets.NewAssets().Nora).Wait(Config.DefaultTimeoutDuration())).To(Exit(0))

		// .NET Apps must be pushed with --no-start --no-route before running cf map-route
		appName3 = random_name.CATSRandomName("APP")
		Expect(cf.Cf("push",
			appName3,
			"--no-start",
			"--no-route",
			"-s", Config.GetWindowsStack(),
			"-b", Config.GetHwcBuildpackName(),
			"-m", DEFAULT_MEMORY_LIMIT,
			"-p", assets.NewAssets().Nora).Wait(Config.DefaultTimeoutDuration())).To(Exit(0))

		hostname = appName1

		Expect(cf.Cf("map-route", appName2, domain, "--hostname", hostname, "--path", app2Path).Wait(Config.DefaultTimeoutDuration())).To(Exit(0))
		Expect(cf.Cf("map-route", appName3, domain, "--hostname", hostname, "--path", app3Path).Wait(Config.DefaultTimeoutDuration())).To(Exit(0))

		Expect(cf.Cf("start", appName2).Wait(Config.CfPushTimeoutDuration())).To(Exit(0))
		Expect(cf.Cf("start", appName3).Wait(Config.CfPushTimeoutDuration())).To(Exit(0))
	})

	AfterEach(func() {
		app_helpers.AppReport(appName1, Config.DefaultTimeoutDuration())
		app_helpers.AppReport(appName2, Config.DefaultTimeoutDuration())
		app_helpers.AppReport(appName3, Config.DefaultTimeoutDuration())

		Expect(cf.Cf("delete", appName1, "-f", "-r").Wait(Config.DefaultTimeoutDuration())).To(Exit(0))
		Expect(cf.Cf("delete", appName2, "-f", "-r").Wait(Config.DefaultTimeoutDuration())).To(Exit(0))
		Expect(cf.Cf("delete", appName3, "-f", "-r").Wait(Config.DefaultTimeoutDuration())).To(Exit(0))
	})

	Context("when another app has a route with a context path", func() {
		It("routes to app with context path", func() {
			Eventually(func() string {
				return helpers.CurlAppRoot(Config, hostname)
			}, Config.DefaultTimeoutDuration()).Should(ContainSubstring(strings.ToLower(appName1)))

			Eventually(func() string {
				return helpers.CurlApp(Config, hostname, fmt.Sprintf("%s/env/VCAP_APPLICATION", app2Path))
			}, Config.DefaultTimeoutDuration()).Should(ContainSubstring(fmt.Sprintf(`\"application_name\":\"%s\"`, appName2)))

			Eventually(func() string {
				return helpers.CurlApp(Config, hostname, fmt.Sprintf("%s/env/VCAP_APPLICATION", app3Path))
			}, Config.DefaultTimeoutDuration()).Should(ContainSubstring(fmt.Sprintf(`\"application_name\":\"%s\"`, appName3)))
		})
	})
})
