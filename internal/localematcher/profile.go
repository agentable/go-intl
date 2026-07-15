package localematcher

import "slices"

type languageMatchingProfile struct {
	paradigmLocales []string
	matchVariables  []languageMatchVariable
	rules           []languageMatchRule
}

type languageMatchVariable struct {
	name            string
	sourceRegions   []string
	expandedRegions []string
}

type languageMatchRule struct {
	desired   string
	supported string
	distance  int
	oneWay    bool
}

func defaultLanguageMatchingProfile() languageMatchingProfile {
	profile := languageMatchingProfile{
		paradigmLocales: slices.Clone(generatedLanguageMatchingProfile.paradigmLocales),
		matchVariables:  make([]languageMatchVariable, len(generatedLanguageMatchingProfile.matchVariables)),
		rules:           slices.Clone(generatedLanguageMatchingProfile.rules),
	}
	for i, variable := range generatedLanguageMatchingProfile.matchVariables {
		profile.matchVariables[i] = languageMatchVariable{
			name:            variable.name,
			sourceRegions:   slices.Clone(variable.sourceRegions),
			expandedRegions: slices.Clone(variable.expandedRegions),
		}
	}
	return profile
}
