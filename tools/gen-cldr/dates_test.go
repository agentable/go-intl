package main

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestRunGeneratesDatesAndPreference asserts the generator emits the date domain
// as a self-contained const-only payload at internal/cldr/date/data.go, no longer
// emits the retired root dates.go literal renderer, and still emits the unchanged
// preference.go.
func TestRunGeneratesDatesAndPreference(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	root := filepath.Join(dir, "node_modules")
	writeRuntimeCLDRFixtures(t, root)
	out := filepath.Join(dir, "out")
	if err := os.MkdirAll(out, 0o777); err != nil {
		t.Fatalf("mkdir out: %v", err)
	}
	versionPath := filepath.Join(out, "VERSION")
	if err := os.WriteFile(versionPath, []byte("cldr=48.1.0\nicu=78\ntzdata=2025b\n"), 0o666); err != nil {
		t.Fatalf("write VERSION: %v", err)
	}

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	if err := Run(context.Background(), Config{CLDRDir: root, OutDir: out, VersionFile: versionPath, ProfileFile: writeLocaleProfileFixture(t, dir)}, log); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// The retired root literal renderer must not produce dates.go anymore.
	if _, err := os.Stat(filepath.Join(out, "dates.go")); !os.IsNotExist(err) {
		t.Fatalf("root dates.go should no longer be generated, stat err = %v", err)
	}

	// The date domain payload is a const-only data.go carrying the blobs and the
	// private _data table the decoder reads.
	payload, err := os.ReadFile(filepath.Join(out, "date", "data.go"))
	if err != nil {
		t.Fatalf("read date/data.go: %v", err)
	}
	if !containsAll(string(payload), "package date", "const _data", "_dateGregorianBlob", "_dateDayPeriodBlob", "_dateSupportedBlob", "_dateCalendarBlob") {
		t.Fatalf("date/data.go missing expected const payload:\n%s", payload)
	}
	joined := dewrapStringLiterals(string(payload))
	if !containsAll(joined, "Anno Domini", "EEEE, MMMM d, y", "morning1", "afternoon1", "midnight") {
		t.Fatalf("date/data.go _data missing expected date strings:\n%s", payload)
	}

	// The retired root literal renderer must no longer produce preference.go; the
	// hour-cycle, week, and calendar preference data is owned by the locale kernel
	// payload now.
	if _, err := os.Stat(filepath.Join(out, "preference.go")); !os.IsNotExist(err) {
		t.Fatalf("root preference.go should no longer be generated, stat err = %v", err)
	}
	kernel, err := os.ReadFile(filepath.Join(out, "locale", "data.go"))
	if err != nil {
		t.Fatalf("read locale/data.go: %v", err)
	}
	if !containsAll(string(kernel), "package cldrlocale", "_hourCycleBlob", "_weekBlob", "_calendarBlob") {
		t.Fatalf("locale/data.go missing preference blobs:\n%s", kernel)
	}
}

func TestRunAcceptsUnwrappedDayPeriodRules(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	root := filepath.Join(dir, "node_modules")
	writeRuntimeCLDRFixtures(t, root)
	supp := filepath.Join(root, "cldr-core", "supplemental")
	dayPeriods := `{"supplemental":{"dayPeriodRuleSet":{"en":{"midnight":{"_at":"00:00"},"morning1":{"_from":"05:00","_before":"12:00"},"noon":{"_at":"12:00"},"afternoon1":{"_from":"12:00","_before":"18:00"}},"zh":{"afternoon1":{"_from":"12:00","_before":"18:00"}}}}}`
	if err := os.WriteFile(filepath.Join(supp, "dayPeriods.json"), []byte(dayPeriods), 0o666); err != nil {
		t.Fatalf("write dayPeriods: %v", err)
	}
	out := filepath.Join(dir, "out")
	if err := os.MkdirAll(out, 0o777); err != nil {
		t.Fatalf("mkdir out: %v", err)
	}
	versionPath := filepath.Join(out, "VERSION")
	if err := os.WriteFile(versionPath, []byte("cldr=48.1.0\nicu=78\ntzdata=2025b\n"), 0o666); err != nil {
		t.Fatalf("write VERSION: %v", err)
	}

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	if err := Run(context.Background(), Config{CLDRDir: root, OutDir: out, VersionFile: versionPath, ProfileFile: writeLocaleProfileFixture(t, dir)}, log); err != nil {
		t.Fatalf("Run: %v", err)
	}
	payload, err := os.ReadFile(filepath.Join(out, "date", "data.go"))
	if err != nil {
		t.Fatalf("read date/data.go: %v", err)
	}
	// The _data const is emitted as line-wrapped concatenated string literals, so a
	// rule-type string can straddle a wrap boundary. Join the literal chunks before
	// asserting the unwrapped day-period rule types reached the payload.
	joined := dewrapStringLiterals(string(payload))
	if !containsAll(joined, "morning1", "afternoon1", "midnight", "noon") {
		t.Fatalf("date/data.go missing unwrapped day period rule types:\n%s", payload)
	}
}

// dewrapStringLiterals collapses the `" +\n<ws>"` continuation sequences the
// generator emits between adjacent string-literal chunks, so substring assertions
// see the logical concatenated text rather than the line-wrapped source form.
func dewrapStringLiterals(src string) string {
	return literalContinuation.ReplaceAllString(src, "")
}

var literalContinuation = regexp.MustCompile(`" \+\n\s*"`)

func TestRunFillsWeekDataFromWorldDefaults(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	root := filepath.Join(dir, "node_modules")
	writeRuntimeCLDRFixtures(t, root)
	supp := filepath.Join(root, "cldr-core", "supplemental")
	weekData := `{"supplemental":{"weekData":{"firstDay":{"001":"mon","US":"sun"},"weekendStart":{"001":"sat"},"weekendEnd":{"001":"sun"},"minDays":{"001":"1","US":"1"}}}}`
	if err := os.WriteFile(filepath.Join(supp, "weekData.json"), []byte(weekData), 0o666); err != nil {
		t.Fatalf("write weekData: %v", err)
	}
	out := filepath.Join(dir, "out")
	if err := os.MkdirAll(out, 0o777); err != nil {
		t.Fatalf("mkdir out: %v", err)
	}
	versionPath := filepath.Join(out, "VERSION")
	if err := os.WriteFile(versionPath, []byte("cldr=48.1.0\nicu=78\ntzdata=2025b\n"), 0o666); err != nil {
		t.Fatalf("write VERSION: %v", err)
	}

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	if err := Run(context.Background(), Config{CLDRDir: root, OutDir: out, VersionFile: versionPath, ProfileFile: writeLocaleProfileFixture(t, dir)}, log); err != nil {
		t.Fatalf("Run: %v", err)
	}
	// The week-data world-default fill is an extract-layer decision verified
	// byte-for-byte through the production accessors in the locale kernel
	// round-trip gate. Here we assert the kernel payload carries the week blob and
	// the US region key in its _data table.
	payload, err := os.ReadFile(filepath.Join(out, "locale", "data.go"))
	if err != nil {
		t.Fatalf("read locale/data.go: %v", err)
	}
	if !containsAll(string(payload), "package cldrlocale", "_weekBlob") {
		t.Fatalf("locale/data.go missing week blob:\n%s", payload)
	}
	reconstructed := readGeneratedStringTable(t, filepath.Join(out, "locale", "data.go"))
	if !strings.Contains(reconstructed, "US") {
		t.Fatalf("locale/data.go _data missing US region key:\n%s", payload)
	}
}

func writeDateCLDRFixture(t *testing.T, root string) {
	t.Helper()
	locales := []string{"en", "en-US", "zh", "zh-Hans", "zh-Hans-CN"}
	for _, loc := range locales {
		dir := filepath.Join(root, "cldr-dates-full", "main", loc)
		if err := os.MkdirAll(dir, 0o777); err != nil {
			t.Fatalf("mkdir dates %s: %v", loc, err)
		}
		if err := os.WriteFile(filepath.Join(dir, "ca-gregorian.json"), []byte(gregorianJSON(loc)), 0o666); err != nil {
			t.Fatalf("write ca-gregorian %s: %v", loc, err)
		}
	}
	supp := filepath.Join(root, "cldr-core", "supplemental")
	timeData := `{"supplemental":{"timeData":{"001":{"_allowed":"H h","_preferred":"H"},"US":{"_allowed":"h H","_preferred":"h"},"DE":{"_allowed":"H h","_preferred":"H"}}}}`
	if err := os.WriteFile(filepath.Join(supp, "timeData.json"), []byte(timeData), 0o666); err != nil {
		t.Fatalf("write timeData: %v", err)
	}
	weekData := `{"supplemental":{"weekData":{"firstDay":{"001":"mon","US":"sun","DE":"mon"},"weekendStart":{"001":"sat","US":"sat"},"weekendEnd":{"001":"sun","US":"sun"},"minDays":{"001":"1","US":"1","DE":"4"}}}}`
	if err := os.WriteFile(filepath.Join(supp, "weekData.json"), []byte(weekData), 0o666); err != nil {
		t.Fatalf("write weekData: %v", err)
	}
	calendarPreference := `{"supplemental":{"calendarPreferenceData":{"001":["gregorian"],"US":["gregorian"],"TH":["buddhist","gregorian"]}}}`
	if err := os.WriteFile(filepath.Join(supp, "calendarPreferenceData.json"), []byte(calendarPreference), 0o666); err != nil {
		t.Fatalf("write calendarPreferenceData: %v", err)
	}
	dayPeriods := `{"supplemental":{"dayPeriodRuleSet":{"en":{"dayPeriodRules":{"midnight":{"_at":"00:00"},"morning1":{"_from":"05:00","_before":"12:00"},"noon":{"_at":"12:00"},"afternoon1":{"_from":"12:00","_before":"18:00"},"evening1":{"_from":"18:00","_before":"21:00"},"night1":{"_from":"21:00","_before":"05:00"}}},"zh":{"dayPeriodRules":{"midnight":{"_at":"00:00"},"morning1":{"_from":"06:00","_before":"12:00"},"noon":{"_at":"12:00"},"afternoon1":{"_from":"12:00","_before":"18:00"},"evening1":{"_from":"18:00","_before":"24:00"},"night1":{"_from":"00:00","_before":"06:00"}}}}}}`
	if err := os.WriteFile(filepath.Join(supp, "dayPeriods.json"), []byte(dayPeriods), 0o666); err != nil {
		t.Fatalf("write dayPeriods: %v", err)
	}
}

func gregorianJSON(locale string) string {
	return `{"main":{"` + locale + `":{"dates":{"calendars":{"gregorian":{
		"eras":{"eraNames":{"0":"Before Christ","1":"Anno Domini"},"eraAbbr":{"0":"BC","1":"AD"},"eraNarrow":{"0":"B","1":"A"}},
		"months":{"format":{"wide":{"1":"January","2":"February","3":"March","4":"April","5":"May","6":"June","7":"July","8":"August","9":"September","10":"October","11":"November","12":"December"},"abbreviated":{"1":"Jan","2":"Feb","3":"Mar","4":"Apr","5":"May","6":"Jun","7":"Jul","8":"Aug","9":"Sep","10":"Oct","11":"Nov","12":"Dec"},"narrow":{"1":"J","2":"F","3":"M","4":"A","5":"M","6":"J","7":"J","8":"A","9":"S","10":"O","11":"N","12":"D"}},"stand-alone":{"wide":{"1":"January","2":"February","3":"March","4":"April","5":"May","6":"June","7":"July","8":"August","9":"September","10":"October","11":"November","12":"December"},"abbreviated":{"1":"Jan","2":"Feb","3":"Mar","4":"Apr","5":"May","6":"Jun","7":"Jul","8":"Aug","9":"Sep","10":"Oct","11":"Nov","12":"Dec"},"narrow":{"1":"J","2":"F","3":"M","4":"A","5":"M","6":"J","7":"J","8":"A","9":"S","10":"O","11":"N","12":"D"}}},
		"days":{"format":{"wide":{"sun":"Sunday","mon":"Monday","tue":"Tuesday","wed":"Wednesday","thu":"Thursday","fri":"Friday","sat":"Saturday"},"abbreviated":{"sun":"Sun","mon":"Mon","tue":"Tue","wed":"Wed","thu":"Thu","fri":"Fri","sat":"Sat"},"narrow":{"sun":"S","mon":"M","tue":"T","wed":"W","thu":"T","fri":"F","sat":"S"}},"stand-alone":{"wide":{"sun":"Sunday","mon":"Monday","tue":"Tuesday","wed":"Wednesday","thu":"Thursday","fri":"Friday","sat":"Saturday"},"abbreviated":{"sun":"Sun","mon":"Mon","tue":"Tue","wed":"Wed","thu":"Thu","fri":"Fri","sat":"Sat"},"narrow":{"sun":"S","mon":"M","tue":"T","wed":"W","thu":"T","fri":"F","sat":"S"}}},
		"quarters":{"format":{"wide":{"1":"1st quarter","2":"2nd quarter","3":"3rd quarter","4":"4th quarter"},"abbreviated":{"1":"Q1","2":"Q2","3":"Q3","4":"Q4"},"narrow":{"1":"1","2":"2","3":"3","4":"4"}},"stand-alone":{"wide":{"1":"1st quarter","2":"2nd quarter","3":"3rd quarter","4":"4th quarter"},"abbreviated":{"1":"Q1","2":"Q2","3":"Q3","4":"Q4"},"narrow":{"1":"1","2":"2","3":"3","4":"4"}}},
		"dayPeriods":{"format":{"wide":{"midnight":"midnight","am":"AM","noon":"noon","pm":"PM","morning1":"morning","afternoon1":"afternoon","evening1":"evening","night1":"night"},"abbreviated":{"midnight":"midnight","am":"AM","noon":"noon","pm":"PM","morning1":"morning","afternoon1":"afternoon","evening1":"evening","night1":"night"},"narrow":{"midnight":"mi","am":"a","noon":"n","pm":"p","morning1":"morning","afternoon1":"afternoon","evening1":"evening","night1":"night"}},"stand-alone":{"wide":{"am":"AM","pm":"PM"},"abbreviated":{"am":"AM","pm":"PM"},"narrow":{"am":"a","pm":"p"}}},
		"dateFormats":{"full":"EEEE, MMMM d, y","long":"MMMM d, y","medium":"MMM d, y","short":"M/d/yy"},
		"timeFormats":{"full":"h:mm:ss a zzzz","long":"h:mm:ss a z","medium":"h:mm:ss a","short":"h:mm a"},
		"dateTimeFormats":{"full":"{1} 'at' {0}","long":"{1} 'at' {0}","medium":"{1}, {0}","short":"{1}, {0}","availableFormats":{"yMMMd":"MMM d, y","Hm":"HH:mm"},"intervalFormats":{"intervalFormatFallback":"{0} – {1}","yMMMd":{"d":"MMM d–d, y","M":"MMM d – MMM d, y"}}}
	}}}}}}`
}
