package matching

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/aarondl/null/v8"
)

// CSV header constants - Niina's "test group v1" format
const (
	// Core fields
	HeaderEmail      = "email" // optional - if provided, used instead of generated
	HeaderFirstName  = "first_name"
	HeaderLastName   = "last_name"
	HeaderAddress    = "address"
	HeaderBirthday   = "birthday"
	HeaderGender     = "gender"
	HeaderHeightCM   = "height_cm"
	HeaderLatitude  = "latitude"
	HeaderLongitude = "longitude"

	// Dating preferences
	HeaderDatingPrefAgeRangeStart = "dating_pref_age_range_start"
	HeaderDatingPrefAgeRangeEnd   = "dating_pref_age_range_end"
	HeaderDatingPreference        = "dating_preference" // singular: Male, Female, Any

	// Qualitative profile fields
	HeaderSelfPortrait               = "self_portrait"
	HeaderInterests                  = "interests"
	HeaderWellbeingHabits            = "wellbeing_habits"
	HeaderSelfCareHabits             = "self_care_habits"
	HeaderMoneyManagement            = "money_management"
	HeaderSelfReflectionCapabilities = "self_reflection_capabilities"
	HeaderMoralFrameworks            = "moral_frameworks"
	HeaderLifeGoals                  = "life_goals"
	HeaderPartnershipValues          = "partnership_values"
	HeaderMutualCommitment           = "mutual_commitment"
	HeaderSpiritualityGrowthMindset  = "spirituality_growth_mindset"
	HeaderCulturalValues             = "cultural_values"
	HeaderFamilyPlanning             = "family_planning"
	HeaderIdealDate                  = "ideal_date"
	HeaderRedGreenFlags              = "red_green_flags"

	// Quantitative profile fields (0-1 floats)
	HeaderExtroversionSocialEnergy = "extroversion_social_energy"
	HeaderRoutineVsSpontaneity     = "routine_vs_spontaneity"
	HeaderAgreeableness            = "agreeableness"
	HeaderConscientiousness        = "conscientiousness"
	HeaderNeuroticism              = "neuroticism"
	HeaderDominanceLevel           = "dominance_level"
	HeaderEmotionalExpressiveness  = "emotional_expressiveness"
	HeaderSexDrive                 = "sex_drive"
	HeaderGeographicalMobility     = "geographical_mobility"

	// Categorical profile fields
	HeaderConflictResolutionStyle = "conflict_resolution_style"
	HeaderSexualityPreferences    = "sexuality_preferences"
	HeaderReligion                = "religion"
)

// Required headers for CSV validation - Niina's flat format
var requiredHeaders = []string{
	HeaderFirstName,
	HeaderLastName,
	HeaderBirthday,
	HeaderGender,
	HeaderHeightCM,
	HeaderLatitude,
	HeaderLongitude,
	HeaderDatingPreference,
	// Qualitative
	HeaderSelfPortrait,
	HeaderInterests,
	HeaderWellbeingHabits,
	HeaderSelfCareHabits,
	HeaderMoneyManagement,
	HeaderSelfReflectionCapabilities,
	HeaderMoralFrameworks,
	HeaderLifeGoals,
	HeaderPartnershipValues,
	HeaderMutualCommitment,
	HeaderSpiritualityGrowthMindset,
	HeaderCulturalValues,
	HeaderFamilyPlanning,
	HeaderIdealDate,
	HeaderRedGreenFlags,
	// Quantitative
	HeaderExtroversionSocialEnergy,
	HeaderRoutineVsSpontaneity,
	HeaderAgreeableness,
	HeaderConscientiousness,
	HeaderNeuroticism,
	HeaderDominanceLevel,
	HeaderEmotionalExpressiveness,
	HeaderSexDrive,
	HeaderGeographicalMobility,
	// Categorical
	HeaderConflictResolutionStyle,
	HeaderSexualityPreferences,
	HeaderReligion,
}

// PopulationRow represents a single row from the CSV file.
// Uses Niina's "test group v1" flat column format.
type PopulationRow struct {
	// Core fields
	Email     string    `json:"email"` // optional - if provided, used instead of generated
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Address   string    `json:"address"`   // optional
	Birthday  time.Time `json:"birthday"`  // parsed from YYYY-MM-DD format
	Age       int       `json:"age"`       // calculated from birthday
	Gender    string    `json:"gender"`    // Male, Female, Non-binary
	HeightCM  int       `json:"height_cm"` // height in centimeters
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`

	// Dating preferences
	DatingPrefAgeRangeStart int      `json:"dating_pref_age_range_start"` // optional
	DatingPrefAgeRangeEnd   int      `json:"dating_pref_age_range_end"`   // optional
	DatingPreference        string   `json:"dating_preference"`           // Male, Female, Any
	DatingPreferences       []string `json:"dating_preferences"`          // expanded from DatingPreference

	// Profile details (parsed from flat columns)
	ProfileDetails *ProfileDetails `json:"profile_details"`
}

// ProfileDetails contains the qualitative and quantitative profile data.
// Uses null.* types to match aipgmodel.Profile directly (no conversion needed).
type ProfileDetails struct {
	// Qualitative fields
	WellbeingHabits            null.String `json:"wellbeing_habits"`
	Interests                  null.String `json:"interests"`
	SenseOfHumor               null.String `json:"sense_of_humor"`
	SelfCareHabits             null.String `json:"self_care_habits"`
	MoneyManagement            null.String `json:"money_management"`
	SelfReflectionCapabilities null.String `json:"self_reflection_capabilities"`
	MoralFrameworks            null.String `json:"moral_frameworks"`
	LifeGoals                  null.String `json:"life_goals"`
	PartnershipValues          null.String `json:"partnership_values"`
	SpiritualityGrowthMindset  null.String `json:"spirituality_growth_mindset"`
	CulturalValues             null.String `json:"cultural_values"`
	FamilyPlanning             null.String `json:"family_planning"`
	SelfPortrait               null.String `json:"self_portrait"`
	MutualCommitment           null.String `json:"mutual_commitment"`
	IdealDate                  null.String `json:"ideal_date"`
	RedGreenFlags              null.String `json:"red_green_flags"`

	// Quantitative fields
	ExtroversionSocialEnergy null.Float64 `json:"extroversion_social_energy"`
	RoutineVsSpontaneity     null.Float64 `json:"routine_vs_spontaneity"`
	Agreeableness            null.Float64 `json:"agreeableness"`
	Conscientiousness        null.Float64 `json:"conscientiousness"`
	Neuroticism              null.Float32 `json:"neuroticism"`
	DominanceLevel           null.Float64 `json:"dominance_level"`
	EmotionalExpressiveness  null.Float64 `json:"emotional_expressiveness"`
	SexDrive                 null.Float64 `json:"sex_drive"`
	GeographicalMobility     null.Float64 `json:"geographical_mobility"`

	// Categorical fields
	ConflictResolutionStyle null.String `json:"conflict_resolution_style"`
	SexualityPreferences    null.String `json:"sexuality_preferences"`
	Religion                null.String `json:"religion"`
}

// PopulationCSVResult holds the result of parsing a CSV file.
type PopulationCSVResult struct {
	Rows           []PopulationRow `json:"rows"`
	TotalRows      int             `json:"total_rows"`
	ValidRows      int             `json:"valid_rows"`
	SkippedRows    int             `json:"skipped_rows"`
	ParsingErrors  []string        `json:"parsing_errors,omitempty"`
	ValidationErrs []string        `json:"validation_errors,omitempty"`
}

// PopulateResult represents the result of populating users across all databases.
type PopulateResult struct {
	TotalProcessed    int      `json:"total_processed"`
	BackendAppUsers   int      `json:"backend_app_users_created"`
	AIBackendProfiles int      `json:"ai_backend_profiles_created"`
	SupabaseAuthUsers int      `json:"supabase_auth_users_created"`
	DatingPreferences int      `json:"dating_preferences_created"`
	Errors            []string `json:"errors,omitempty"`

	// Cleanup counts - records deleted before insertion
	CleanedBackendAppUsers   int `json:"cleaned_backend_app_users"`
	CleanedAIBackendProfiles int `json:"cleaned_ai_backend_profiles"`
	CleanedSupabaseAuthUsers int `json:"cleaned_supabase_auth_users"`
}

// CSV parsing errors
var (
	ErrEmptyCSV           = errors.New("csv file is empty")
	ErrMissingHeaders     = errors.New("csv is missing required headers")
	ErrInvalidHeaderCount = errors.New("csv header count does not match expected")
	ErrNoDataRows         = errors.New("csv has no data rows")
)

// ParsePopulationCSV parses a CSV reader and returns parsed population rows.
func ParsePopulationCSV(reader io.Reader) (*PopulationCSVResult, error) {
	csvReader := csv.NewReader(reader)
	csvReader.TrimLeadingSpace = true

	// Read header row
	headers, err := csvReader.Read()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil, ErrEmptyCSV
		}
		return nil, fmt.Errorf("read csv headers: %w", err)
	}

	// Validate headers
	headerMap, err := validateHeaders(headers)
	if err != nil {
		return nil, err
	}

	result := &PopulationCSVResult{
		Rows:           make([]PopulationRow, 0),
		ParsingErrors:  make([]string, 0),
		ValidationErrs: make([]string, 0),
	}

	rowNum := 1 // 1-indexed for human readability (header is row 1)
	for {
		rowNum++
		record, err := csvReader.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			result.ParsingErrors = append(result.ParsingErrors,
				fmt.Sprintf("row %d: failed to read: %v", rowNum, err))
			result.SkippedRows++
			continue
		}

		result.TotalRows++

		row, parseErr := parseRow(record, headerMap, rowNum)
		if parseErr != nil {
			result.ParsingErrors = append(result.ParsingErrors, parseErr.Error())
			result.SkippedRows++
			continue
		}

		if validationErr := validateRow(row, rowNum); validationErr != nil {
			result.ValidationErrs = append(result.ValidationErrs, validationErr.Error())
			result.SkippedRows++
			continue
		}

		result.Rows = append(result.Rows, *row)
		result.ValidRows++
	}

	if result.TotalRows == 0 {
		return nil, ErrNoDataRows
	}

	return result, nil
}

// validateHeaders validates and maps CSV headers to their indices.
func validateHeaders(headers []string) (map[string]int, error) {
	headerMap := make(map[string]int)

	// Normalize and map headers
	for i, h := range headers {
		normalized := strings.ToLower(strings.TrimSpace(h))
		headerMap[normalized] = i
	}

	// Check for required headers
	var missing []string
	for _, required := range requiredHeaders {
		if _, ok := headerMap[required]; !ok {
			missing = append(missing, required)
		}
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("%w: %s", ErrMissingHeaders, strings.Join(missing, ", "))
	}

	return headerMap, nil
}

// parseRow parses a single CSV row into a PopulationRow.
// Uses Niina's flat column format with birthday and profile fields as separate columns.
func parseRow(record []string, headerMap map[string]int, rowNum int) (*PopulationRow, error) {
	getValue := func(header string) string {
		if idx, ok := headerMap[header]; ok && idx < len(record) {
			return strings.TrimSpace(record[idx])
		}
		return ""
	}

	// Parse birthday (YYYY-MM-DD format)
	birthdayStr := getValue(HeaderBirthday)
	birthday, err := time.Parse("2006-01-02", birthdayStr)
	if err != nil {
		return nil, fmt.Errorf("row %d: invalid birthday '%s' (expected YYYY-MM-DD): %w", rowNum, birthdayStr, err)
	}

	// Calculate age from birthday
	age := calculateAge(birthday)

	// Parse height_cm
	heightStr := getValue(HeaderHeightCM)
	height, err := strconv.Atoi(heightStr)
	if err != nil {
		return nil, fmt.Errorf("row %d: invalid height_cm '%s': %w", rowNum, heightStr, err)
	}

	// Parse latitude
	latStr := getValue(HeaderLatitude)
	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		return nil, fmt.Errorf("row %d: invalid latitude '%s': %w", rowNum, latStr, err)
	}

	// Parse longitude
	lonStr := getValue(HeaderLongitude)
	lon, err := strconv.ParseFloat(lonStr, 64)
	if err != nil {
		return nil, fmt.Errorf("row %d: invalid longitude '%s': %w", rowNum, lonStr, err)
	}

	// Parse dating_preference (singular: Male, Female, Any)
	datingPref := getValue(HeaderDatingPreference)
	datingPrefs := expandDatingPreference(datingPref)

	// Parse optional dating_pref_age_range_start and dating_pref_age_range_end
	var datingPrefAgeStart, datingPrefAgeEnd int
	if ageStartStr := getValue(HeaderDatingPrefAgeRangeStart); ageStartStr != "" {
		datingPrefAgeStart, err = strconv.Atoi(ageStartStr)
		if err != nil {
			return nil, fmt.Errorf("row %d: invalid dating_pref_age_range_start '%s': %w", rowNum, ageStartStr, err)
		}
	}
	if ageEndStr := getValue(HeaderDatingPrefAgeRangeEnd); ageEndStr != "" {
		datingPrefAgeEnd, err = strconv.Atoi(ageEndStr)
		if err != nil {
			return nil, fmt.Errorf("row %d: invalid dating_pref_age_range_end '%s': %w", rowNum, ageEndStr, err)
		}
	}

	// Parse quantitative fields (0-1 floats)
	extroversion, err := parseFloat64(getValue(HeaderExtroversionSocialEnergy), rowNum, HeaderExtroversionSocialEnergy)
	if err != nil {
		return nil, err
	}
	routine, err := parseFloat64(getValue(HeaderRoutineVsSpontaneity), rowNum, HeaderRoutineVsSpontaneity)
	if err != nil {
		return nil, err
	}
	agreeableness, err := parseFloat64(getValue(HeaderAgreeableness), rowNum, HeaderAgreeableness)
	if err != nil {
		return nil, err
	}
	conscientiousness, err := parseFloat64(getValue(HeaderConscientiousness), rowNum, HeaderConscientiousness)
	if err != nil {
		return nil, err
	}
	neuroticism, err := parseFloat64(getValue(HeaderNeuroticism), rowNum, HeaderNeuroticism)
	if err != nil {
		return nil, err
	}
	dominance, err := parseFloat64(getValue(HeaderDominanceLevel), rowNum, HeaderDominanceLevel)
	if err != nil {
		return nil, err
	}
	emotional, err := parseFloat64(getValue(HeaderEmotionalExpressiveness), rowNum, HeaderEmotionalExpressiveness)
	if err != nil {
		return nil, err
	}
	sexDrive, err := parseFloat64(getValue(HeaderSexDrive), rowNum, HeaderSexDrive)
	if err != nil {
		return nil, err
	}
	geoMobility, err := parseFloat64(getValue(HeaderGeographicalMobility), rowNum, HeaderGeographicalMobility)
	if err != nil {
		return nil, err
	}

	// Build ProfileDetails from flat columns
	profileDetails := &ProfileDetails{
		// Qualitative fields
		SelfPortrait:               null.StringFrom(getValue(HeaderSelfPortrait)),
		Interests:                  null.StringFrom(getValue(HeaderInterests)),
		WellbeingHabits:            null.StringFrom(getValue(HeaderWellbeingHabits)),
		SelfCareHabits:             null.StringFrom(getValue(HeaderSelfCareHabits)),
		MoneyManagement:            null.StringFrom(getValue(HeaderMoneyManagement)),
		SelfReflectionCapabilities: null.StringFrom(getValue(HeaderSelfReflectionCapabilities)),
		MoralFrameworks:            null.StringFrom(getValue(HeaderMoralFrameworks)),
		LifeGoals:                  null.StringFrom(getValue(HeaderLifeGoals)),
		PartnershipValues:          null.StringFrom(getValue(HeaderPartnershipValues)),
		MutualCommitment:           null.StringFrom(getValue(HeaderMutualCommitment)),
		SpiritualityGrowthMindset:  null.StringFrom(getValue(HeaderSpiritualityGrowthMindset)),
		CulturalValues:             null.StringFrom(getValue(HeaderCulturalValues)),
		FamilyPlanning:             null.StringFrom(getValue(HeaderFamilyPlanning)),
		IdealDate:                  null.StringFrom(getValue(HeaderIdealDate)),
		RedGreenFlags:              null.StringFrom(getValue(HeaderRedGreenFlags)),

		// Quantitative fields (0-1 floats)
		ExtroversionSocialEnergy: null.Float64From(extroversion),
		RoutineVsSpontaneity:     null.Float64From(routine),
		Agreeableness:            null.Float64From(agreeableness),
		Conscientiousness:        null.Float64From(conscientiousness),
		Neuroticism:              null.Float32From(float32(neuroticism)),
		DominanceLevel:           null.Float64From(dominance),
		EmotionalExpressiveness:  null.Float64From(emotional),
		SexDrive:                 null.Float64From(sexDrive),
		GeographicalMobility:     null.Float64From(geoMobility),

		// Categorical fields
		ConflictResolutionStyle: null.StringFrom(getValue(HeaderConflictResolutionStyle)),
		SexualityPreferences:    null.StringFrom(getValue(HeaderSexualityPreferences)),
		Religion:                null.StringFrom(getValue(HeaderReligion)),
	}

	row := &PopulationRow{
		Email:                   getValue(HeaderEmail), // optional
		FirstName:               getValue(HeaderFirstName),
		LastName:                getValue(HeaderLastName),
		Address:                 getValue(HeaderAddress),
		Birthday:                birthday,
		Age:                     age,
		Gender:                  getValue(HeaderGender),
		HeightCM:                height,
		Latitude:                lat,
		Longitude:               lon,
		DatingPrefAgeRangeStart: datingPrefAgeStart,
		DatingPrefAgeRangeEnd:   datingPrefAgeEnd,
		DatingPreference:        datingPref,
		DatingPreferences:       datingPrefs,
		ProfileDetails:          profileDetails,
	}

	return row, nil
}

// parseFloat64 parses a string to float64 with a helpful error message.
func parseFloat64(s string, rowNum int, fieldName string) (float64, error) {
	if s == "" {
		return 0, fmt.Errorf("row %d: %s is required", rowNum, fieldName)
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("row %d: invalid %s '%s': %w", rowNum, fieldName, s, err)
	}
	return v, nil
}

// calculateAge calculates age from a birthday as of today.
func calculateAge(birthday time.Time) int {
	now := time.Now()
	age := now.Year() - birthday.Year()
	// Adjust if birthday hasn't occurred yet this year
	if now.YearDay() < birthday.YearDay() {
		age--
	}
	return age
}

// expandDatingPreference expands a singular dating preference value to a list.
// "Male" -> ["Male"], "Female" -> ["Female"], "Any" -> ["Male", "Female", "Non-Binary"]
func expandDatingPreference(pref string) []string {
	switch pref {
	case "Male":
		return []string{"Male"}
	case "Female":
		return []string{"Female"}
	case "Any":
		return []string{"Male", "Female", "Non-Binary"}
	default:
		// Unknown preference - return as-is in a slice
		if pref != "" {
			return []string{pref}
		}
		return nil
	}
}

// Valid dating preference values
var validDatingPreferences = map[string]bool{
	"Male":   true,
	"Female": true,
	"Any":    true,
}

// validateRow validates a parsed row for business rules.
func validateRow(row *PopulationRow, rowNum int) error {
	if row.FirstName == "" {
		return fmt.Errorf("row %d: first_name is required", rowNum)
	}

	if row.LastName == "" {
		return fmt.Errorf("row %d: last_name is required", rowNum)
	}

	if row.Age < 18 || row.Age > 120 {
		return fmt.Errorf("row %d: age (calculated from birthday) must be between 18 and 120, got %d", rowNum, row.Age)
	}

	validGenders := map[string]bool{"Male": true, "Female": true, "Non-binary": true}
	if !validGenders[row.Gender] {
		return fmt.Errorf("row %d: invalid gender '%s', must be Male, Female, or Non-binary", rowNum, row.Gender)
	}

	if !validDatingPreferences[row.DatingPreference] {
		return fmt.Errorf("row %d: invalid dating_preference '%s', must be Male, Female, or Any", rowNum, row.DatingPreference)
	}

	if row.HeightCM < 100 || row.HeightCM > 250 {
		return fmt.Errorf("row %d: height_cm must be between 100 and 250, got %d", rowNum, row.HeightCM)
	}

	if row.Latitude < -90 || row.Latitude > 90 {
		return fmt.Errorf("row %d: latitude must be between -90 and 90", rowNum)
	}

	if row.Longitude < -180 || row.Longitude > 180 {
		return fmt.Errorf("row %d: longitude must be between -180 and 180", rowNum)
	}

	if row.ProfileDetails == nil {
		return fmt.Errorf("row %d: profile details are required", rowNum)
	}

	// Validate quantitative fields are in 0-1 range
	if row.ProfileDetails.ExtroversionSocialEnergy.Valid {
		v := row.ProfileDetails.ExtroversionSocialEnergy.Float64
		if v < 0 || v > 1 {
			return fmt.Errorf("row %d: extroversion_social_energy must be between 0 and 1, got %f", rowNum, v)
		}
	}
	if row.ProfileDetails.RoutineVsSpontaneity.Valid {
		v := row.ProfileDetails.RoutineVsSpontaneity.Float64
		if v < 0 || v > 1 {
			return fmt.Errorf("row %d: routine_vs_spontaneity must be between 0 and 1, got %f", rowNum, v)
		}
	}
	if row.ProfileDetails.Agreeableness.Valid {
		v := row.ProfileDetails.Agreeableness.Float64
		if v < 0 || v > 1 {
			return fmt.Errorf("row %d: agreeableness must be between 0 and 1, got %f", rowNum, v)
		}
	}
	if row.ProfileDetails.Conscientiousness.Valid {
		v := row.ProfileDetails.Conscientiousness.Float64
		if v < 0 || v > 1 {
			return fmt.Errorf("row %d: conscientiousness must be between 0 and 1, got %f", rowNum, v)
		}
	}
	if row.ProfileDetails.Neuroticism.Valid {
		v := row.ProfileDetails.Neuroticism.Float32
		if v < 0 || v > 1 {
			return fmt.Errorf("row %d: neuroticism must be between 0 and 1, got %f", rowNum, v)
		}
	}
	if row.ProfileDetails.DominanceLevel.Valid {
		v := row.ProfileDetails.DominanceLevel.Float64
		if v < 0 || v > 1 {
			return fmt.Errorf("row %d: dominance_level must be between 0 and 1, got %f", rowNum, v)
		}
	}
	if row.ProfileDetails.EmotionalExpressiveness.Valid {
		v := row.ProfileDetails.EmotionalExpressiveness.Float64
		if v < 0 || v > 1 {
			return fmt.Errorf("row %d: emotional_expressiveness must be between 0 and 1, got %f", rowNum, v)
		}
	}
	if row.ProfileDetails.SexDrive.Valid {
		v := row.ProfileDetails.SexDrive.Float64
		if v < 0 || v > 1 {
			return fmt.Errorf("row %d: sex_drive must be between 0 and 1, got %f", rowNum, v)
		}
	}
	if row.ProfileDetails.GeographicalMobility.Valid {
		v := row.ProfileDetails.GeographicalMobility.Float64
		if v < 0 || v > 1 {
			return fmt.Errorf("row %d: geographical_mobility must be between 0 and 1, got %f", rowNum, v)
		}
	}

	return nil
}
