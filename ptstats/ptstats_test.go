/*
Package ptstats implements an interceptor to update patient statistics
for a given county or county subdivision (town).

Carlton Duffett
*/
package ptstats

import (
	"github.com/intervention-engine/fhir/models"
	"github.com/stretchr/testify/assert"
	"testing"
)

type CountyStats struct {
	Population, PopulationMale, PopulationFemale int64
	PopulationPerSquareMile                      float64
}

type CousubStats struct {
	Population, PopulationMale, PopulationFemale int64
	PopulationPerSquareMile                      float64
}

// TestCountySubdivisionDataAccess implements the CountySubdivisionDataAccess
// interface without a database connection for testing purposes only.
type TestCountySubdivisionDataAccess struct{}

func (da *TestCountySubdivisionDataAccess) GetCountySubdivisionFp(city string) string {

	if city == "" {
		return ""
	}

	switch city {
	case "Boston":
		return "07000"
	case "Bedford":
		return "04615"
	default:
		return "00000" // undefined subdivision
	}
}

func (da *TestCountySubdivisionDataAccess) GetCountyFp(cousubFp string) string {

	if cousubFp == "" {
		return ""
	}

	switch cousubFp {
	case "07000": // Boston
		return "025"
	case "04615": // Bedford
		return "017"
	default:
		return "001" // Barnstable County
	}
}

func (da *TestCountySubdivisionDataAccess) GetStateFp(countyFp string) string {

	if countyFp == "" {
		return ""
	}

	return "025" // Massachusetts
}

// TestSyntheticCountyStatsData is a collection of data items for a single county
type TestSyntheticCountyStatsData struct {
	Pop, MalePop, FemalePop int64
	SqMiles, PopPerSqMile   float64
}

func (data *TestSyntheticCountyStatsData) ModifyPopulationCount(countyFp string, maleDelta, femaleDelta int64) {
	data.Pop += (maleDelta + femaleDelta)
	data.MalePop += maleDelta
	data.FemalePop += femaleDelta
	data.PopPerSqMile = float64(data.Pop) / data.SqMiles
}

// TestSyntheticCountyStatsDataAccess implements the SyntheticCountyStatsDataAccess
// interface without a database connection for testing purposes only.
type TestSyntheticCountyStatsDataAccess struct {
	data map[string]*TestSyntheticCountyStatsData
}

func (da *TestSyntheticCountyStatsDataAccess) GetPopulation(countyFp string) int64 {
	return da.data[countyFp].Pop
}

func (da *TestSyntheticCountyStatsDataAccess) GetMalePopulation(countyFp string) int64 {
	return da.data[countyFp].MalePop
}

func (da *TestSyntheticCountyStatsDataAccess) GetFemalePopulation(countyFp string) int64 {
	return da.data[countyFp].FemalePop
}

func (da *TestSyntheticCountyStatsDataAccess) GetPopulationPerSquareMile(countyFp string) float64 {
	return da.data[countyFp].PopPerSqMile
}

func (da *TestSyntheticCountyStatsDataAccess) AddMale(countyFp string) {
	da.data[countyFp].ModifyPopulationCount(countyFp, 1, 0)
}

func (da *TestSyntheticCountyStatsDataAccess) AddFemale(countyFp string) {
	da.data[countyFp].ModifyPopulationCount(countyFp, 0, 1)
}

func (da *TestSyntheticCountyStatsDataAccess) RemoveMale(countyFp string) {
	da.data[countyFp].ModifyPopulationCount(countyFp, -1, 0)
}

func (da *TestSyntheticCountyStatsDataAccess) RemoveFemale(countyFp string) {
	da.data[countyFp].ModifyPopulationCount(countyFp, 0, -1)
}

// TestSyntheticCountySubdivisionData is a collection of data items for a single county subdivision
type TestSyntheticCountySubdivisionStatsData struct {
	Pop, MalePop, FemalePop int64
	SqMiles, PopPerSqMile   float64
}

func (data *TestSyntheticCountySubdivisionStatsData) ModifyPopulationCount(countyFp, cousubFp string, maleDelta, femaleDelta int64) {
	data.Pop += (maleDelta + femaleDelta)
	data.MalePop += maleDelta
	data.FemalePop += femaleDelta
	data.PopPerSqMile = float64(data.Pop) / data.SqMiles
}

// TestSyntheticCountySubdivisionStatsDataAccess implements the SyntheticCountySubdivisonStatsDataAccess
// interface without a database connection for testing purposes only.
type TestSyntheticCountySubdivisionStatsDataAccess struct {
	data map[string]*TestSyntheticCountySubdivisionStatsData
}

func (da *TestSyntheticCountySubdivisionStatsDataAccess) GetPopulation(countyFp string, cousubFp string) int64 {
	return da.data[cousubFp].Pop
}

func (da *TestSyntheticCountySubdivisionStatsDataAccess) GetMalePopulation(countyFp string, cousubFp string) int64 {
	return da.data[cousubFp].MalePop
}

func (da *TestSyntheticCountySubdivisionStatsDataAccess) GetFemalePopulation(countyFp string, cousubFp string) int64 {
	return da.data[cousubFp].FemalePop
}

func (da *TestSyntheticCountySubdivisionStatsDataAccess) GetPopulationPerSquareMile(countyFp string, cousubFp string) float64 {
	return da.data[cousubFp].PopPerSqMile
}

func (da *TestSyntheticCountySubdivisionStatsDataAccess) AddMale(countyFp string, cousubFp string) {
	da.data[cousubFp].ModifyPopulationCount(countyFp, cousubFp, 1, 0)
}

func (da *TestSyntheticCountySubdivisionStatsDataAccess) AddFemale(countyFp string, cousubFp string) {
	da.data[cousubFp].ModifyPopulationCount(countyFp, cousubFp, 0, 1)
}

func (da *TestSyntheticCountySubdivisionStatsDataAccess) RemoveMale(countyFp string, cousubFp string) {
	da.data[cousubFp].ModifyPopulationCount(countyFp, cousubFp, -1, 0)
}

func (da *TestSyntheticCountySubdivisionStatsDataAccess) RemoveFemale(countyFp string, cousubFp string) {
	da.data[cousubFp].ModifyPopulationCount(countyFp, cousubFp, 0, -1)
}

// TestPatientStatsCreateInterceptor tests the PatientStatsCreateInterceptor's ability to update
// patient statistics after a new patient is added to the database
func TestPatientStatsCreateInterceptor(t *testing.T) {
	assert := assert.New(t)

	da := createNewPatientStatsDataAccess()

	createInterceptor := &PatientStatsCreateInterceptor{
		DataAccess: da,
	}

	// 1. Test a CreateMale operation
	// ------------------------------------------------------------------------
	patient := createNewPatient("Boston", "male")
	countyFp, cousubFp, err := da.IdentifyCountyAndSubdivision(patient)

	if err != nil {
		panic(err)
	}

	oldCounty := getCountyStatistics(da, countyFp)
	oldCousub := getCousubStatistics(da, countyFp, cousubFp)

	createInterceptor.After(patient)

	newCounty := getCountyStatistics(da, countyFp)
	newCousub := getCousubStatistics(da, countyFp, cousubFp)

	// Test updated county statistics
	assert.Equal(oldCounty.Population+1, newCounty.Population, "County population should increment by 1")
	assert.Equal(oldCounty.PopulationMale+1, newCounty.PopulationMale, "County male population should increment by 1")
	assert.Equal(oldCounty.PopulationFemale, newCounty.PopulationFemale, "County female population should not change")
	newCountyPopPerSqMile := float64(oldCounty.Population+1) / 5.0 // 5.0 sqMiles
	assert.Equal(newCountyPopPerSqMile, newCounty.PopulationPerSquareMile, "County population density should change with an increase in population")

	// Test updated cousub statistics
	assert.Equal(oldCousub.Population+1, newCousub.Population, "Subdivision population should increment by 1")
	assert.Equal(oldCousub.PopulationMale+1, newCousub.PopulationMale, "Subdivision male population should increment by 1")
	assert.Equal(oldCousub.PopulationFemale, newCousub.PopulationFemale, "Subdivision female population should not change")
	newCousubPopPerSqMile := float64(oldCousub.Population+1) / 5.0
	assert.Equal(newCousubPopPerSqMile, newCousub.PopulationPerSquareMile, "Subdivision population density should change with an increase in population")

	// 2. Test a CreateFemale operation
	// ------------------------------------------------------------------------
	patient.Gender = "female"

	oldCounty = newCounty
	oldCousub = newCousub

	createInterceptor.After(patient)

	newCounty = getCountyStatistics(da, countyFp)
	newCousub = getCousubStatistics(da, countyFp, cousubFp)

	// Test updated county statistics
	assert.Equal(oldCounty.Population+1, newCounty.Population, "County population should increment by 1")
	assert.Equal(oldCounty.PopulationMale, newCounty.PopulationMale, "County male population should not change")
	assert.Equal(oldCounty.PopulationFemale+1, newCounty.PopulationFemale, "County female population should increment by 1")
	newCountyPopPerSqMile = float64(oldCounty.Population+1) / 5.0
	assert.Equal(newCountyPopPerSqMile, newCounty.PopulationPerSquareMile, "County population density should change with an increase in population")

	// Test updated cousub statistics
	assert.Equal(oldCousub.Population+1, newCousub.Population, "Subdivision population should increment by 1")
	assert.Equal(oldCousub.PopulationMale, newCousub.PopulationMale, "Subdivision male population should not change")
	assert.Equal(oldCousub.PopulationFemale+1, newCousub.PopulationFemale, "Subdivision female population should increment by 1")
	newCousubPopPerSqMile = float64(oldCousub.Population+1) / 5.0
	assert.Equal(newCousubPopPerSqMile, newCousub.PopulationPerSquareMile, "Subdivision population density should change with an increase in population")
}

// TestPatientStatsUpdateInterceptor tests the PatientStatsUpdateInterceptor's ability to update
// patient statistics after a new patient is added to the database
func TestPatientStatsUpdateInterceptor(t *testing.T) {
	assert := assert.New(t)

	da := createNewPatientStatsDataAccess()

	updateInterceptor := &PatientStatsUpdateInterceptor{
		DataAccess: da,
	}

	// 3. Test a UpdateMale operation with no relevant changes
	// ------------------------------------------------------------------------
	patient := createNewPatient("Boston", "male")
	countyFp, cousubFp, err := da.IdentifyCountyAndSubdivision(patient)

	if err != nil {
		panic(err)
	}

	oldCounty := getCountyStatistics(da, countyFp)
	oldCousub := getCousubStatistics(da, countyFp, cousubFp)

	updateInterceptor.Before(patient)
	updateInterceptor.After(patient)

	newCounty := getCountyStatistics(da, countyFp)
	newCousub := getCousubStatistics(da, countyFp, cousubFp)

	testStatsDontChange(t, oldCounty, newCounty, oldCousub, newCousub)

	// 4. Test a UpdateFemale operation with no relevant changes
	// ------------------------------------------------------------------------
	patient.Gender = "female"

	oldCounty = getCountyStatistics(da, countyFp)
	oldCousub = getCousubStatistics(da, countyFp, cousubFp)

	updateInterceptor.Before(patient)
	updateInterceptor.After(patient)

	newCounty = getCountyStatistics(da, countyFp)
	newCousub = getCousubStatistics(da, countyFp, cousubFp)

	testStatsDontChange(t, oldCounty, newCounty, oldCousub, newCousub)

	// 5. Test a UpdateMale operation with relevant statistic changes
	// ------------------------------------------------------------------------
	patient.Gender = "male"

	oldCounty = getCountyStatistics(da, countyFp)
	oldCousub = getCousubStatistics(da, countyFp, cousubFp)

	updateInterceptor.Before(patient)
	afterPatient := createNewPatient("Bedford", "male")
	updateInterceptor.After(afterPatient)

	newCounty = getCountyStatistics(da, countyFp)
	newCousub = getCousubStatistics(da, countyFp, cousubFp)

	// we test here for a decrease in Boston's cousub and county populations. There
	// is no need to create a separate data layer to manage Bedford's statistics,
	// since we already verified that we can add a new male to the data layer in
	// TestPatientStatsCreateInterceptor.

	// Test updated county statistics
	assert.Equal(oldCounty.Population-1, newCounty.Population, "County population should decrement by 1")
	assert.Equal(oldCounty.PopulationMale-1, newCounty.PopulationMale, "County male population should decrement by 1")
	assert.Equal(oldCounty.PopulationFemale, newCounty.PopulationFemale, "County female population should not change")
	newCountyPopPerSqMile := float64(oldCounty.Population-1) / 5.0
	assert.Equal(newCountyPopPerSqMile, newCounty.PopulationPerSquareMile, "County population density should change with an decrease in population")

	// Test updated cousub statistics
	assert.Equal(oldCousub.Population-1, newCousub.Population, "Subdivision population should decrement by 1")
	assert.Equal(oldCousub.PopulationMale-1, newCousub.PopulationMale, "Subdivision male population should decrement by 1")
	assert.Equal(oldCousub.PopulationFemale, newCousub.PopulationFemale, "Subdivision female population should not change")
	newCousubPopPerSqMile := float64(oldCousub.Population-1) / 5.0
	assert.Equal(newCousubPopPerSqMile, newCousub.PopulationPerSquareMile, "Subdivision population density should change with an decrease in population")

	// 6. Test a UpdateFemale operation with relevant statistic changes
	// ------------------------------------------------------------------------
	patient.Gender = "female"

	oldCounty = getCountyStatistics(da, countyFp)
	oldCousub = getCousubStatistics(da, countyFp, cousubFp)

	updateInterceptor.Before(patient)
	afterPatient.Gender = "female"
	updateInterceptor.After(afterPatient)

	newCounty = getCountyStatistics(da, countyFp)
	newCousub = getCousubStatistics(da, countyFp, cousubFp)

	// Test updated county statistics
	assert.Equal(oldCounty.Population-1, newCounty.Population, "County population should decrement by 1")
	assert.Equal(oldCounty.PopulationMale, newCounty.PopulationMale, "County male population should not change")
	assert.Equal(oldCounty.PopulationFemale-1, newCounty.PopulationFemale, "County female population should decrement by 1")
	newCountyPopPerSqMile = float64(oldCounty.Population-1) / 5.0
	assert.Equal(newCountyPopPerSqMile, newCounty.PopulationPerSquareMile, "County population density should change with an decrease in population")

	// Test updated cousub statistics
	assert.Equal(oldCousub.Population-1, newCousub.Population, "Subdivision population should decrement by 1")
	assert.Equal(oldCousub.PopulationMale, newCousub.PopulationMale, "Subdivision male population should not change")
	assert.Equal(oldCousub.PopulationFemale-1, newCousub.PopulationFemale, "Subdivision female population should decrement by 1")
	newCousubPopPerSqMile = float64(oldCousub.Population-1) / 5.0
	assert.Equal(newCousubPopPerSqMile, newCousub.PopulationPerSquareMile, "Subdivision population density should change with an decrease in population")
}

// TestPatientStatsDeleteInterceptor tests the PatientStatsDeleteInterceptor's ability to update
// patient statistics after a new patient is added to the database
func TestPatientStatsDeleteInterceptor(t *testing.T) {
	assert := assert.New(t)

	da := createNewPatientStatsDataAccess()

	deleteInterceptor := &PatientStatsDeleteInterceptor{
		DataAccess: da,
	}

	// 7. Test a DeleteMale operation
	// ------------------------------------------------------------------------
	patient := createNewPatient("Boston", "male")
	countyFp, cousubFp, err := da.IdentifyCountyAndSubdivision(patient)

	if err != nil {
		panic(err)
	}

	oldCounty := getCountyStatistics(da, countyFp)
	oldCousub := getCousubStatistics(da, countyFp, cousubFp)

	deleteInterceptor.After(patient)

	newCounty := getCountyStatistics(da, countyFp)
	newCousub := getCousubStatistics(da, countyFp, cousubFp)

	// Test updated county statistics
	assert.Equal(oldCounty.Population-1, newCounty.Population, "County population should decrement by 1")
	assert.Equal(oldCounty.PopulationMale-1, newCounty.PopulationMale, "County male population should decrement by 1")
	assert.Equal(oldCounty.PopulationFemale, newCounty.PopulationFemale, "County female population should not change")
	newCountyPopPerSqMile := float64(oldCounty.Population-1) / 5.0
	assert.Equal(newCountyPopPerSqMile, newCounty.PopulationPerSquareMile, "County population density should change with an decrease in population")

	// Test updated cousub statistics
	assert.Equal(oldCousub.Population-1, newCousub.Population, "Subdivision population should decrement by 1")
	assert.Equal(oldCousub.PopulationMale-1, newCousub.PopulationMale, "Subdivision male population should decrement by 1")
	assert.Equal(oldCousub.PopulationFemale, newCousub.PopulationFemale, "Subdivision female population should not change")
	newCousubPopPerSqMile := float64(oldCousub.Population-1) / 5.0
	assert.Equal(newCousubPopPerSqMile, newCousub.PopulationPerSquareMile, "Subdivision population density should change with an decrease in population")

	// 8. Test a DeleteFemale operation
	// ------------------------------------------------------------------------
	patient.Gender = "female"

	oldCounty = newCounty
	oldCousub = newCousub

	deleteInterceptor.After(patient)

	newCounty = getCountyStatistics(da, countyFp)
	newCousub = getCousubStatistics(da, countyFp, cousubFp)

	// Test updated county statistics
	assert.Equal(oldCounty.Population-1, newCounty.Population, "County population should decrement by 1")
	assert.Equal(oldCounty.PopulationMale, newCounty.PopulationMale, "County male population should not change")
	assert.Equal(oldCounty.PopulationFemale-1, newCounty.PopulationFemale, "County female population should decrement by 1")
	newCountyPopPerSqMile = float64(oldCounty.Population-1) / 5.0
	assert.Equal(newCountyPopPerSqMile, newCounty.PopulationPerSquareMile, "County population density should change with an decrease in population")

	// Test updated cousub statistics
	assert.Equal(oldCousub.Population-1, newCousub.Population, "Subdivision population should decrement by 1")
	assert.Equal(oldCousub.PopulationMale, newCousub.PopulationMale, "Subdivision male population should not change")
	assert.Equal(oldCousub.PopulationFemale-1, newCousub.PopulationFemale, "Subdivision female population should decrement by 1")
	newCousubPopPerSqMile = float64(oldCousub.Population-1) / 5.0
	assert.Equal(newCousubPopPerSqMile, newCousub.PopulationPerSquareMile, "Subdivision population density should change with an decrease in population")
}

// TestPatientStatsInterceptorErrorHandling tests that the interceptors and underlying
// data access layers handle errors as expected
func TestPatientStatsInterceptorErrorHandling(t *testing.T) {
	// 9. Test an invalid gender (fails silently)
	// ------------------------------------------------------------------------

	// 10. Test a city that does not exist (fails silently)
	// ------------------------------------------------------------------------
}

// testStatsDontChange is a reusable testing submodule that verifies no patient statistics have changed
func testStatsDontChange(t *testing.T, oldCounty, newCounty CountyStats, oldCousub, newCousub CousubStats) {

	// Test updated county statistics
	assert.Equal(t, oldCounty.Population, newCounty.Population, "County population should not change")
	assert.Equal(t, oldCounty.PopulationMale, newCounty.PopulationMale, "County male population should not change")
	assert.Equal(t, oldCounty.PopulationFemale, newCounty.PopulationFemale, "County female population should not change")
	assert.Equal(t, oldCounty.PopulationPerSquareMile, newCounty.PopulationPerSquareMile, "County population density should not change")

	// Test updated cousub statistics
	assert.Equal(t, oldCousub.Population, newCousub.Population, "Subdivision population should not change")
	assert.Equal(t, oldCousub.PopulationMale, newCousub.PopulationMale, "Subdivision male population should not change")
	assert.Equal(t, oldCousub.PopulationFemale, newCousub.PopulationFemale, "Subdivision female population should not change")
	assert.Equal(t, oldCousub.PopulationPerSquareMile, newCousub.PopulationPerSquareMile, "Subdivision population density should not change")
}

// getCountyStatistics gets the county statistics that are currently in the database
func getCountyStatistics(da *PatientStatsDataAccess, countyFp string) CountyStats {

	county := CountyStats{
		Population:              da.CountyStats.GetPopulation(countyFp),
		PopulationMale:          da.CountyStats.GetMalePopulation(countyFp),
		PopulationFemale:        da.CountyStats.GetFemalePopulation(countyFp),
		PopulationPerSquareMile: da.CountyStats.GetPopulationPerSquareMile(countyFp),
	}
	return county
}

// getCousubStatistics gets the county subdivision statistics that are currently in the database
func getCousubStatistics(da *PatientStatsDataAccess, countyFp, cousubFp string) CousubStats {

	cousub := CousubStats{
		Population:              da.CousubStats.GetPopulation(countyFp, cousubFp),
		PopulationMale:          da.CousubStats.GetMalePopulation(countyFp, cousubFp),
		PopulationFemale:        da.CousubStats.GetFemalePopulation(countyFp, cousubFp),
		PopulationPerSquareMile: da.CousubStats.GetPopulationPerSquareMile(countyFp, cousubFp),
	}
	return cousub
}

// createNewPatientStatsDataAccess creates a minimally populated
// PatientStatsDataAccess interface
func createNewPatientStatsDataAccess() *PatientStatsDataAccess {

	// Initialize county data
	countyData := make(map[string]*TestSyntheticCountyStatsData)
	countyData["025"] = &TestSyntheticCountyStatsData{ // 025 = Suffolk County
		Pop:          100,
		MalePop:      50,
		FemalePop:    50,
		SqMiles:      5.0,
		PopPerSqMile: 20.0,
	}
	countyData["017"] = &TestSyntheticCountyStatsData{ // 017 = Middlesex County
		Pop:          100,
		MalePop:      50,
		FemalePop:    50,
		SqMiles:      5.0,
		PopPerSqMile: 20.0,
	}

	// Initialize cousub data
	cousubData := make(map[string]*TestSyntheticCountySubdivisionStatsData)
	cousubData["07000"] = &TestSyntheticCountySubdivisionStatsData{ // 07000 = Boston
		Pop:          100,
		MalePop:      50,
		FemalePop:    50,
		SqMiles:      5.0,
		PopPerSqMile: 20.0,
	}
	cousubData["04615"] = &TestSyntheticCountySubdivisionStatsData{ // 04615 = Bedford
		Pop:          100,
		MalePop:      50,
		FemalePop:    50,
		SqMiles:      5.0,
		PopPerSqMile: 20.0,
	}

	return &PatientStatsDataAccess{
		CountyStats: &TestSyntheticCountyStatsDataAccess{data: countyData},
		CousubStats: &TestSyntheticCountySubdivisionStatsDataAccess{data: cousubData},
		Cousub:      &TestCountySubdivisionDataAccess{},
	}
}

// createNewPatient creates a minimally populated Patient object for testing
func createNewPatient(city, gender string) *models.Patient {

	address := models.Address{
		City:  city,
		State: "MA",
	}

	return &models.Patient{
		Address: []models.Address{address},
		Gender:  gender,
	}
}
