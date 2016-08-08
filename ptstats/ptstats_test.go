/*
Package ptstats implements an interceptor to update patient statistics
for a given county or county subdivision (town).

Carlton Duffett
*/
package ptstats

import (
	"bytes"
	"io/ioutil"
	"log"

	"github.com/gin-gonic/gin"
)

// TestCountySubdivisionDataAcess implements the CountySubdivisionDataAccess
// interface without a database connection for testing purposes only.
type TestCountySubdivisionDataAccess struct{}

func (da TestCountySubdivisionDataAccess) GetCountySubdivisionFp(city string) string {
	switch city {
	case "Boston":
		return "07000"
	case "Bedford":
		return "04615"
	default:
		return "00000" // undefined
	}
}

func (da TestCountySubdivisionDataAccess) GetCountyFp(cousubFp string) string {
	switch cousubFp {
	case "07000": // Boston
		return "025"
	case "04615": // Bedford
		return "017"
	default:
		return "000" // undefined
	}
}

func (da TestCountySubdivisionDataAccess) GetStateFp(countyFp string) string {
	return "025" // Massachusetts
}

// TestSyntheticCountyStatsDataAccess implements the SyntheticCountyStatsDataAccess
// interface without a database connection for testing purposes only.
type TestSyntheticCountyStatsDataAccess struct {
	pop, malePop, femalePop int64
	sqMiles, popPerSqMile   float64
}

func (da TestSyntheticCountyStatsDataAccess) GetPopulation(countyFp string) int64 {
	return da.pop
}

func (da TestSyntheticCountyStatsDataAccess) GetMalePopulation(countyFp string) int64 {
	return da.malePop
}

func (da TestSyntheticCountyStatsDataAccess) GetFemalePopulation(countyFp string) int64 {
	return da.femalePop
}

func (da TestSyntheticCountyStatsDataAccess) GetPopulationPerSquareMile(countyFp string) float64 {
	return da.popPerSqMile
}

func (da TestSyntheticCountyStatsDataAccess) AddMale(countyFp string) {
	da.malePop += 1
	da.pop += 1
	da.updatePopPerSqMile()
}

func (da TestSyntheticCountyStatsDataAccess) AddFemale(countyFp string) {
	da.femalePop += 1
	da.pop += 1
	da.updatePopPerSqMile()
}

func (da TestSyntheticCountyStatsDataAccess) RemoveMale(countyFp string) {
	da.malePop -= 1
	da.pop -= 1
	da.updatePopPerSqMile()
}

func (da TestSyntheticCountyStatsDataAccess) RemoveFemale(countyFp string) {
	da.femalePop -= 1
	da.pop -= 1
	da.updatePopPerSqMile()
}

func (da TestSyntheticCountyStatsDataAccess) updatePopPerSqMile() {
	da.popPerSqMile = float64(da.pop / da.sqMiles)
}

// TestSyntheticCountySubdivisionStatsDataAccess implements the SyntheticCountySubdivisonStatsDataAccess
// interface without a database connection for testing purposes only.
type TestSyntheticCountySubdivisionStatsDataAccess struct {
	pop, malePop, femalePop int64
	sqMiles, popPerSqMile   float64
}

func (da TestSyntheticCountySubdivisionStatsDataAccess) GetPopulation(cousubFp string) int64 {
	return da.pop
}

func (da TestSyntheticCountySubdivisionStatsDataAccess) GetMalePopulation(cousubFp string) int64 {
	return da.malePop
}

func (da TestSyntheticCountySubdivisionStatsDataAccess) GetFemalePopulation(cousubFp string) int64 {
	return da.femalePop
}

func (da TestSyntheticCountySubdivisionStatsDataAccess) GetPopulationPerSquareMile(cousubFp string) float64 {
	return da.popPerSqMile
}

func (da TestSyntheticCountySubdivisionStatsDataAccess) AddMale(cousubFp string) {
	da.malePop += 1
	da.pop += 1
	da.updatePopPerSqMile()
}

func (da TestSyntheticCountySubdivisionStatsDataAccess) AddFemale(cousubFp string) {
	da.femalePop += 1
	da.pop += 1
	da.updatePopPerSqMile()
}

func (da TestSyntheticCountySubdivisionStatsDataAccess) RemoveMale(cousubFp string) {
	da.malePop -= 1
	da.pop -= 1
	da.updatePopPerSqMile()
}

func (da TestSyntheticCountySubdivisionStatsDataAccess) RemoveFemale(cousubFp string) {
	da.femalePop -= 1
	da.pop -= 1
	da.updatePopPerSqMile()
}

func (da TestSyntheticCountySubdivisionStatsDataAccess) updatePopPerSqMile() {
	da.popPerSqMile = float64(da.pop / da.sqMiles)
}
