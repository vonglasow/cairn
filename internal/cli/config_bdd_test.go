package cli

// Executable acceptance scenarios for bdd/config.feature: where cairn finds
// the file that says its two places, and what happens when it says nothing,
// says something wrong, or is not there at all.

import (
	"testing"

	"github.com/cucumber/godog"
)

func initConfig(t *testing.T) func(*godog.ScenarioContext) {
	return func(sc *godog.ScenarioContext) {
		w := newWorld(t)
		w.register(sc)
		w.registerConfig(sc)
	}
}

func TestAcceptance_Config_TheFileIsFoundWithoutBeingNamed(t *testing.T) {
	runAcceptance(t, "config.feature", initConfig(t), "@config-is-found-where-configuration-lives")
}

func TestAcceptance_Config_NothingSaidPutsTheStoreUnderTheHome(t *testing.T) {
	runAcceptance(t, "config.feature", initConfig(t), "@config-nothing-said-puts-the-store-under-the-home")
}

func TestAcceptance_Config_TheEnvironmentOverridesTheFileForTheStore(t *testing.T) {
	runAcceptance(t, "config.feature", initConfig(t), "@config-the-environment-overrides-the-file-for-the-store")
}

func TestAcceptance_Config_ThatIsNotTOMLIsRefused(t *testing.T) {
	runAcceptance(t, "config.feature", initConfig(t), "@config-that-is-not-toml-is-refused")
}

func TestAcceptance_Config_AnUnknownNameIsRefused(t *testing.T) {
	runAcceptance(t, "config.feature", initConfig(t), "@config-an-unknown-name-is-refused")
}

func TestAcceptance_Config_APlaceThatIsNotAbsoluteIsRefused(t *testing.T) {
	runAcceptance(t, "config.feature", initConfig(t), "@config-a-place-that-is-not-absolute-is-refused")
}

func TestAcceptance_Config_AStoreThatIsNotAbsoluteIsRefused(t *testing.T) {
	runAcceptance(t, "config.feature", initConfig(t), "@config-a-store-that-is-not-absolute-is-refused")
}

func TestAcceptance_Config_ThatIsNotThereIsTheOrdinaryCase(t *testing.T) {
	runAcceptance(t, "config.feature", initConfig(t), "@config-that-is-not-there-is-the-ordinary-case")
}

func TestAcceptance_Config_ShowNamesTheFileItReads(t *testing.T) {
	runAcceptance(t, "config.feature", initConfig(t), "@config-show-names-the-file-it-reads")
}

func TestAcceptance_Config_ShowSaysWhatNamedEachPlace(t *testing.T) {
	runAcceptance(t, "config.feature", initConfig(t), "@config-show-says-what-named-each-place")
}

func TestAcceptance_Config_ShowSaysWhenNothingNamedAPlace(t *testing.T) {
	runAcceptance(t, "config.feature", initConfig(t), "@config-show-says-when-nothing-named-a-place")
}

func TestAcceptance_Config_ShowTravelsInTheJSON(t *testing.T) {
	runAcceptance(t, "config.feature", initConfig(t), "@config-show-travels-in-the-json")
}

func TestAcceptance_Config_ShowReportsWhatItCannotMakeSenseOf(t *testing.T) {
	runAcceptance(t, "config.feature", initConfig(t), "@config-show-reports-what-it-cannot-make-sense-of")
}

func TestAcceptance_Config_ShowAnswersAnAgentAboutAFileItCannotMakeSenseOf(t *testing.T) {
	runAcceptance(t, "config.feature", initConfig(t), "@config-show-answers-an-agent-about-a-file-it-cannot-make-sense-of")
}
