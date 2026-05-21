package codegen

import (
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/agentable/go-intl/tools/gen-cldr/cldr"
	"github.com/agentable/go-intl/tools/gen-cldr/extract"
)

func renderMetazones(data extract.Metazones, table *StringTable) ([]byte, error) {
	var b strings.Builder
	b.WriteString("package cldr\n\n")
	b.WriteString("import (\n\t\"strconv\"\n\t\"strings\"\n\t\"sync\"\n)\n\n")
	b.WriteString("type metazoneNames struct{ longGeneric, longStandard, longDaylight, shortGeneric, shortStandard, shortDaylight sliceRef }\n\n")
	b.WriteString("type TimeZoneFormats struct{ GMTFormat, GMTZeroFormat, HourFormat string }\n\n")
	b.WriteString("type timeZoneFormatRefs struct{ gmtFormat, gmtZeroFormat, hourFormat sliceRef }\n\n")
	b.WriteString("type metazonePeriod struct {\n\tmetazone sliceRef\n\tstart    int64\n\tend      int64\n}\n\n")
	b.WriteString("type TimeZoneName string\n\n")
	b.WriteString("const (\n\tTimeZoneNameShort TimeZoneName = \"short\"\n\tTimeZoneNameLong TimeZoneName = \"long\"\n\tTimeZoneNameShortOffset TimeZoneName = \"shortOffset\"\n\tTimeZoneNameLongOffset TimeZoneName = \"longOffset\"\n\tTimeZoneNameShortGeneric TimeZoneName = \"shortGeneric\"\n\tTimeZoneNameLongGeneric TimeZoneName = \"longGeneric\"\n)\n\n")
	b.WriteString(`var (
	metazoneDataOnce        sync.Once
	zoneToMetazones         map[string][]metazonePeriod
	metazoneNamesByLocale   map[Locale]map[string]metazoneNames
	timeZoneNamesByLocale   map[Locale]map[string]metazoneNames
	exemplarCitiesByLocale  map[Locale]map[string]sliceRef
	timeZoneFormatsByLocale map[Locale]timeZoneFormatRefs
)

func zoneMetazoneData() map[string][]metazonePeriod {
	metazoneDataOnce.Do(loadMetazoneData)
	return zoneToMetazones
}

func metazoneNameDataByLocale() map[Locale]map[string]metazoneNames {
	metazoneDataOnce.Do(loadMetazoneData)
	return metazoneNamesByLocale
}

func timeZoneNameDataByLocale() map[Locale]map[string]metazoneNames {
	metazoneDataOnce.Do(loadMetazoneData)
	return timeZoneNamesByLocale
}

func exemplarCityDataByLocale() map[Locale]map[string]sliceRef {
	metazoneDataOnce.Do(loadMetazoneData)
	return exemplarCitiesByLocale
}

func timeZoneFormatDataByLocale() map[Locale]timeZoneFormatRefs {
	metazoneDataOnce.Do(loadMetazoneData)
	return timeZoneFormatsByLocale
}

func loadMetazoneData() {
	zoneToMetazones = `)
	b.WriteString(metazonePeriodMapLiteral(data.ZoneToMetazones, table))
	b.WriteString("\n\n")
	b.WriteString("\tmetazoneNamesByLocale = map[Locale]map[string]metazoneNames{\n")
	for _, locale := range sortedLocaleKeys(data.Names) {
		b.WriteString("\tlocaleIndex[")
		b.WriteString(strconv.Quote(locale))
		b.WriteString("]: ")
		b.WriteString(metazoneNamesMapLiteral(data.Names[locale], table))
		b.WriteString(",\n")
	}
	b.WriteString("\t}\n\n")
	b.WriteString("\ttimeZoneNamesByLocale = map[Locale]map[string]metazoneNames{\n")
	for _, locale := range sortedLocaleKeys(data.ZoneNames) {
		b.WriteString("\tlocaleIndex[")
		b.WriteString(strconv.Quote(locale))
		b.WriteString("]: ")
		b.WriteString(metazoneNamesMapLiteral(data.ZoneNames[locale], table))
		b.WriteString(",\n")
	}
	b.WriteString("\t}\n\n")
	b.WriteString("\texemplarCitiesByLocale = map[Locale]map[string]sliceRef{\n")
	for _, locale := range sortedLocaleKeys(data.ExemplarCities) {
		b.WriteString("\tlocaleIndex[")
		b.WriteString(strconv.Quote(locale))
		b.WriteString("]: ")
		b.WriteString(stringRefMapLiteral(data.ExemplarCities[locale], table))
		b.WriteString(",\n")
	}
	b.WriteString("\t}\n\n")
	b.WriteString("\ttimeZoneFormatsByLocale = map[Locale]timeZoneFormatRefs{\n")
	for _, locale := range sortedLocaleKeys(data.Formats) {
		b.WriteString("\tlocaleIndex[")
		b.WriteString(strconv.Quote(locale))
		b.WriteString("]: ")
		b.WriteString(timeZoneFormatsLiteral(data.Formats[locale], table))
		b.WriteString(",\n")
	}
	b.WriteString("\t}\n")
	b.WriteString("}\n\n")
	b.WriteString(`func ZoneToMetazone(zone string) string {
	return TimeZoneMetazone(zone, 0)
}

func TimeZoneMetazone(zone string, instant int64) string {
	for _, period := range zoneMetazoneData()[zone] {
		if instant >= period.start && instant < period.end {
			return period.metazone.string()
		}
	}
	return ""
}

func TimeZoneDisplayName(loc Locale, zone string, form TimeZoneName, isDST bool, instant int64, offsetMs int64) string {
	kind := "long-generic"
	switch form {
	case TimeZoneNameLong:
		if isDST {
			kind = "long-daylight"
		} else {
			kind = "long-standard"
		}
	case TimeZoneNameShort:
		if isDST {
			kind = "short-daylight"
		} else {
			kind = "short-standard"
		}
	case TimeZoneNameShortGeneric:
		kind = "short-generic"
	}
	if name := loc.TimeZoneName(zone, kind); name != "" {
		return name
	}
	if name := loc.MetazoneName(TimeZoneMetazone(zone, instant), kind); name != "" {
		return name
	}
	if city := loc.ExemplarCity(zone); city != "" {
		return "Time in " + city
	}
	return GMTOffsetName(loc, offsetMs, form)
}

func (l Locale) TimeZoneName(zone, kind string) string {
	return timeZoneNameValue(timeZoneNameDataByLocale()[l][zone], kind)
}

func (l Locale) MetazoneName(metazone, kind string) string {
	return timeZoneNameValue(metazoneNameDataByLocale()[l][metazone], kind)
}

func timeZoneNameValue(data metazoneNames, kind string) string {
	switch kind {
	case "long-standard":
		return data.longStandard.string()
	case "long-daylight":
		return data.longDaylight.string()
	case "short-generic":
		return data.shortGeneric.string()
	case "short-standard":
		return data.shortStandard.string()
	case "short-daylight":
		return data.shortDaylight.string()
	default:
		return data.longGeneric.string()
	}
}

func (l Locale) ExemplarCity(zone string) string {
	return exemplarCityDataByLocale()[l][zone].string()
}

func (l Locale) TimeZoneFormats() TimeZoneFormats {
	data := timeZoneFormatDataByLocale()[l]
	formats := TimeZoneFormats{
		GMTFormat:     data.gmtFormat.string(),
		GMTZeroFormat: data.gmtZeroFormat.string(),
		HourFormat:    data.hourFormat.string(),
	}
	if formats.GMTFormat == "" {
		formats.GMTFormat = "GMT{0}"
	}
	if formats.GMTZeroFormat == "" {
		formats.GMTZeroFormat = "GMT"
	}
	if formats.HourFormat == "" {
		formats.HourFormat = "+HH:mm;-HH:mm"
	}
	return formats
}

// GMTOffsetName formats an offset time-zone name using locale GMT patterns.
func GMTOffsetName(loc Locale, offsetMs int64, form TimeZoneName) string {
	formats := loc.TimeZoneFormats()
	long := form == TimeZoneNameLongOffset || form == TimeZoneNameLong
	offset := offsetPattern(formats.HourFormat, offsetMs, long)
	if offset == "" && !long {
		return formats.GMTZeroFormat
	}
	return strings.ReplaceAll(formats.GMTFormat, "{0}", offset)
}

func offsetPattern(hourFormat string, offsetMs int64, long bool) string {
	positive, negative, ok := strings.Cut(hourFormat, ";")
	if !ok {
		positive, negative = "+HH:mm", "-HH:mm"
	}
	pattern := positive
	if offsetMs < 0 {
		pattern = negative
		offsetMs = -offsetMs
	}
	totalMinutes := offsetMs / 60000
	hours := totalMinutes / 60
	minutes := totalMinutes % 60
	if !long {
		if hours == 0 && minutes == 0 {
			return ""
		}
		if minutes == 0 {
			pattern = removeMinuteField(pattern)
		}
		return replaceOffsetFields(pattern, int(hours), int(minutes), false)
	}
	return replaceOffsetFields(pattern, int(hours), int(minutes), true)
}

func removeMinuteField(pattern string) string {
	i := strings.IndexByte(pattern, 'm')
	if i < 0 {
		return pattern
	}
	start := i
	if start > 0 && pattern[start-1] == ':' {
		start--
	}
	end := i
	for end < len(pattern) && pattern[end] == 'm' {
		end++
	}
	return pattern[:start] + pattern[end:]
}

func replaceOffsetFields(pattern string, hours int, minutes int, padded bool) string {
	hour := strconv.Itoa(hours)
	minute := strconv.Itoa(minutes)
	if padded {
		hour = twoDigitASCII(hours)
		minute = twoDigitASCII(minutes)
	}
	out := replaceFieldRun(pattern, 'H', hour)
	out = replaceFieldRun(out, 'm', minute)
	return out
}

func replaceFieldRun(pattern string, field byte, replacement string) string {
	i := strings.IndexByte(pattern, field)
	if i < 0 {
		return pattern
	}
	end := i
	for end < len(pattern) && pattern[end] == field {
		end++
	}
	return pattern[:i] + replacement + pattern[end:]
}

func twoDigitASCII(value int) string {
	if value < 10 {
		return "0" + strconv.Itoa(value)
	}
	return strconv.Itoa(value)
}
`)
	return FormatFile([]byte(b.String()))
}

func metazonePeriodMapLiteral(values map[string][]cldr.MetazonePeriod, table *StringTable) string {
	return stringKeyedMapLiteral("map[string][]metazonePeriod", values, func(value []cldr.MetazonePeriod) string {
		return metazonePeriodSliceLiteral(value, table)
	})
}

func metazonePeriodSliceLiteral(values []cldr.MetazonePeriod, table *StringTable) string {
	if len(values) == 0 {
		return "nil"
	}
	var b strings.Builder
	b.WriteString("[]metazonePeriod{")
	for _, period := range values {
		b.WriteString("{")
		b.WriteString(table.Add(period.Metazone).GoLiteral())
		b.WriteString(", ")
		b.WriteString(strconv.FormatInt(period.Start, 10))
		b.WriteString(", ")
		b.WriteString(strconv.FormatInt(period.End, 10))
		b.WriteString("}, ")
	}
	b.WriteString("}")
	return b.String()
}

func metazoneNamesMapLiteral(values map[string]cldr.MetazoneNames, table *StringTable) string {
	return stringKeyedMapLiteral("map[string]metazoneNames", values, func(value cldr.MetazoneNames) string {
		return metazoneNamesLiteral(value, table)
	})
}

func metazoneNamesLiteral(names cldr.MetazoneNames, table *StringTable) string {
	return "metazoneNames{" +
		refSliceLiteral(names.LongGeneric, table) + ", " +
		refSliceLiteral(names.LongStandard, table) + ", " +
		refSliceLiteral(names.LongDaylight, table) + ", " +
		refSliceLiteral(names.ShortGeneric, table) + ", " +
		refSliceLiteral(names.ShortStandard, table) + ", " +
		refSliceLiteral(names.ShortDaylight, table) + "}"
}

func timeZoneFormatsLiteral(formats cldr.TimeZoneFormats, table *StringTable) string {
	return "timeZoneFormatRefs{" +
		refSliceLiteral(formats.GMTFormat, table) + ", " +
		refSliceLiteral(formats.GMTZeroFormat, table) + ", " +
		refSliceLiteral(formats.HourFormat, table) + "}"
}

func stringRefMapLiteral(values map[string]string, table *StringTable) string {
	return stringKeyedMapLiteral("map[string]sliceRef", values, func(value string) string {
		return refSliceLiteral(value, table)
	})
}

func refSliceLiteral(value string, table *StringTable) string {
	if value == "" {
		return "sliceRef{}"
	}
	return table.Add(value).GoLiteral()
}

func renderUnits(data extract.Units, table *StringTable) ([]byte, error) {
	var b strings.Builder
	b.WriteString("package cldr\n\n")
	b.WriteString("import \"sync\"\n\n")
	b.WriteString(`type unitID uint8
type unitWidth uint8
type unitPlural uint8

const (
	unitWidthLong unitWidth = iota + 1
	unitWidthNarrow
	unitWidthShort
)

const (
	unitPluralFew unitPlural = iota + 1
	unitPluralMany
	unitPluralOne
	unitPluralOther
	unitPluralTwo
	unitPluralZero
)

type unitPatternKey uint32

type unitPatternRecord struct {
	key     unitPatternKey
	pattern sliceRef
}

type compoundUnitPatternRecord struct {
	key     uint32
	pattern sliceRef
}

`)
	b.WriteString("func unitPatternID(unit string) unitID {\n\tswitch unit {\n")
	for i, unit := range sortedUnitPatternNames(data) {
		fmt.Fprintf(&b, "\tcase %q:\n\t\treturn unitID(%d)\n", unit, i+1)
	}
	b.WriteString("\tdefault:\n\t\treturn 0\n\t}\n}\n\n")
	b.WriteString(`func unitWidthID(width string) unitWidth {
	switch width {
	case "", "long":
		return unitWidthLong
	case "narrow":
		return unitWidthNarrow
	case "short":
		return unitWidthShort
	default:
		return 0
	}
}

func unitPluralID(plural string) unitPlural {
	switch plural {
	case "few":
		return unitPluralFew
	case "many":
		return unitPluralMany
	case "one":
		return unitPluralOne
	case "", "other":
		return unitPluralOther
	case "two":
		return unitPluralTwo
	case "zero":
		return unitPluralZero
	default:
		return 0
	}
}

func makeUnitPatternKey(loc Locale, unit unitID, width unitWidth, plural unitPlural) unitPatternKey {
	return unitPatternKey(uint32(loc)<<16 | uint32(unit)<<8 | uint32(width)<<4 | uint32(plural))
}

func makeCompoundUnitPatternKey(loc Locale, width unitWidth) uint32 {
	return uint32(loc)<<4 | uint32(width)
}

`)
	b.WriteString(`var (
	unitsByLocaleOnce sync.Once
	unitPatterns      []unitPatternRecord
	compoundUnitRows  []compoundUnitPatternRecord
	unitLocales       []Locale
)

func unitPatternData() []unitPatternRecord {
	unitsByLocaleOnce.Do(loadUnitData)
	return unitPatterns
}

func compoundUnitPatternData() []compoundUnitPatternRecord {
	unitsByLocaleOnce.Do(loadUnitData)
	return compoundUnitRows
}

func loadUnitData() {
`)
	b.WriteString("\tunitLocales = []Locale{\n")
	for _, locale := range sortedLocaleKeys(data) {
		b.WriteString("\tlocaleIndex[")
		b.WriteString(strconv.Quote(locale))
		b.WriteString("],\n")
	}
	b.WriteString("\t}\n")
	b.WriteString("\tunitPatterns = []unitPatternRecord{\n")
	unitIDs := unitPatternIDs(data)
	for _, row := range unitPatternRows(data) {
		fmt.Fprintf(&b, "\t\t{makeUnitPatternKey(localeIndex[%q], %d, %s, %s), %s},\n",
			row.locale, unitIDs[row.unit], unitWidthConst(row.width), unitPluralConst(row.plural), table.Add(row.pattern).GoLiteral())
	}
	b.WriteString("\t}\n")
	b.WriteString("\tcompoundUnitRows = []compoundUnitPatternRecord{\n")
	for _, row := range compoundUnitPatternRows(data) {
		fmt.Fprintf(&b, "\t\t{makeCompoundUnitPatternKey(localeIndex[%q], %s), %s},\n",
			row.locale, unitWidthConst(row.width), table.Add(row.pattern).GoLiteral())
	}
	b.WriteString("\t}\n")
	b.WriteString("}\n\n")
	b.WriteString(`func (l Locale) UnitPattern(unit, width, plural string) string {
	unitID := unitPatternID(unit)
	widthID := unitWidthID(width)
	pluralID := unitPluralID(plural)
	if unitID == 0 || widthID == 0 || pluralID == 0 {
		return ""
	}
	key := makeUnitPatternKey(l, unitID, widthID, pluralID)
	rows := unitPatternData()
	i := searchUnitPattern(rows, key)
	if i < len(rows) && rows[i].key == key {
		return rows[i].pattern.string()
	}
	return ""
}

func (l Locale) CompoundUnitPattern(width string) string {
	keyWidth := unitWidthID(width)
	if keyWidth == 0 {
		return ""
	}
	key := makeCompoundUnitPatternKey(l, keyWidth)
	rows := compoundUnitPatternData()
	i := searchCompoundUnitPattern(rows, key)
	if i < len(rows) && rows[i].key == key {
		return rows[i].pattern.string()
	}
	return ""
}

func unitSupportedLocaleTags() []string {
	unitsByLocaleOnce.Do(loadUnitData)
	tags := make([]string, 0, len(unitLocales))
	for _, loc := range unitLocales {
		tags = append(tags, localeRecords[loc].tag.string())
	}
	return tags
}

func searchUnitPattern(rows []unitPatternRecord, key unitPatternKey) int {
	low, high := 0, len(rows)
	for low < high {
		mid := int(uint(low+high) >> 1)
		if rows[mid].key < key {
			low = mid + 1
		} else {
			high = mid
		}
	}
	return low
}

func searchCompoundUnitPattern(rows []compoundUnitPatternRecord, key uint32) int {
	low, high := 0, len(rows)
	for low < high {
		mid := int(uint(low+high) >> 1)
		if rows[mid].key < key {
			low = mid + 1
		} else {
			high = mid
		}
	}
	return low
}
`)
	return FormatFile([]byte(b.String()))
}

type unitPatternRow struct {
	locale  string
	unit    string
	width   string
	plural  string
	pattern string
}

type compoundUnitPatternRow struct {
	locale  string
	width   string
	pattern string
}

func sortedUnitPatternNames(data extract.Units) []string {
	seen := map[string]bool{}
	for _, units := range data {
		for unit := range units {
			seen[unit] = true
		}
	}
	return slices.Sorted(maps.Keys(seen))
}

func unitPatternIDs(data extract.Units) map[string]int {
	ids := map[string]int{}
	for i, unit := range sortedUnitPatternNames(data) {
		ids[unit] = i + 1
	}
	return ids
}

func unitPatternRows(data extract.Units) []unitPatternRow {
	var rows []unitPatternRow
	for _, locale := range sortedLocaleKeys(data) {
		for _, unit := range slices.Sorted(maps.Keys(data[locale])) {
			unitData := data[locale][unit]
			for _, width := range unitWidthOrder {
				widthPatterns := unitData.Patterns[width]
				if len(widthPatterns) == 0 {
					continue
				}
				patterns := widthPatterns[unit]
				for _, plural := range unitPluralOrder {
					if pattern := patterns[plural]; pattern != "" {
						rows = append(rows, unitPatternRow{locale: locale, unit: unit, width: width, plural: plural, pattern: pattern})
					}
				}
			}
		}
	}
	return rows
}

func compoundUnitPatternRows(data extract.Units) []compoundUnitPatternRow {
	var rows []compoundUnitPatternRow
	for _, locale := range sortedLocaleKeys(data) {
		compounds := compoundPatternsForLocale(data[locale])
		for _, width := range unitWidthOrder {
			if pattern := compounds[width]; pattern != "" {
				rows = append(rows, compoundUnitPatternRow{locale: locale, width: width, pattern: pattern})
			}
		}
	}
	return rows
}

func compoundPatternsForLocale(units cldr.Units) map[string]string {
	out := map[string]string{}
	for _, unit := range slices.Sorted(maps.Keys(units)) {
		for _, width := range unitWidthOrder {
			if out[width] == "" {
				out[width] = units[unit].Compound[width]
			}
		}
	}
	return out
}

var unitWidthOrder = []string{"long", "narrow", "short"}

var unitPluralOrder = []string{"few", "many", "one", "other", "two", "zero"}

func unitWidthConst(width string) string {
	switch width {
	case "long":
		return "unitWidthLong"
	case "narrow":
		return "unitWidthNarrow"
	case "short":
		return "unitWidthShort"
	default:
		panic("unknown unit width " + width)
	}
}

func unitPluralConst(plural string) string {
	switch plural {
	case "few":
		return "unitPluralFew"
	case "many":
		return "unitPluralMany"
	case "one":
		return "unitPluralOne"
	case "other":
		return "unitPluralOther"
	case "two":
		return "unitPluralTwo"
	case "zero":
		return "unitPluralZero"
	default:
		panic("unknown unit plural " + plural)
	}
}

func nestedStringMap3Literal(values map[string]map[string]map[string]string, table *StringTable) string {
	return stringKeyedMapLiteral("map[string]map[string]map[string]string", values, func(value map[string]map[string]string) string {
		return nestedStringMapLiteral(value, table)
	})
}
