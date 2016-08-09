/*
Package ptstats implements an interceptor to update patient statistics
for a given county or county subdivision (town).

Carlton Duffett
*/
package ptstats

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"testing"
)

// TestCountySubdivisionDataAccess implements the CountySubdivisionDataAccess
// interface without a database connection for testing purposes only.
type TestCountySubdivisionDataAccess struct{}

func (da TestCountySubdivisionDataAccess) GetCountySubdivisionFp(city string) string {

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

func (da TestCountySubdivisionDataAccess) GetCountyFp(cousubFp string) string {

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

func (da TestCountySubdivisionDataAccess) GetStateFp(countyFp string) string {

	if countyFp == "" {
		return ""
	}

	return "025" // Massachusetts
}

// TestSyntheticCountyStatsDataAccess implements the SyntheticCountyStatsDataAccess
// interface without a database connection for testing purposes only.
type TestSyntheticCountyStatsDataAccess struct {
	pop, malePop, femalePop int64
	sqMiles, popPerSqMile   float64
}

func (da *TestSyntheticCountyStatsDataAccess) GetPopulation(countyFp string) int64 {
	return da.pop
}

func (da *TestSyntheticCountyStatsDataAccess) GetMalePopulation(countyFp string) int64 {
	return da.malePop
}

func (da *TestSyntheticCountyStatsDataAccess) GetFemalePopulation(countyFp string) int64 {
	return da.femalePop
}

func (da *TestSyntheticCountyStatsDataAccess) GetPopulationPerSquareMile(countyFp string) float64 {
	return da.popPerSqMile
}

func (da *TestSyntheticCountyStatsDataAccess) AddMale(countyFp string) {
	da.malePop += 1
	da.pop += 1
	da.updatePopPerSqMile()
}

func (da *TestSyntheticCountyStatsDataAccess) AddFemale(countyFp string) {
	da.femalePop += 1
	da.pop += 1
	da.updatePopPerSqMile()
}

func (da *TestSyntheticCountyStatsDataAccess) RemoveMale(countyFp string) {
	da.malePop -= 1
	da.pop -= 1
	da.updatePopPerSqMile()
}

func (da *TestSyntheticCountyStatsDataAccess) RemoveFemale(countyFp string) {
	da.femalePop -= 1
	da.pop -= 1
	da.updatePopPerSqMile()
}

func (da *TestSyntheticCountyStatsDataAccess) updatePopPerSqMile() {
	da.popPerSqMile = float64(da.pop) / da.sqMiles
}

// TestSyntheticCountySubdivisionStatsDataAccess implements the SyntheticCountySubdivisonStatsDataAccess
// interface without a database connection for testing purposes only.
type TestSyntheticCountySubdivisionStatsDataAccess struct {
	pop, malePop, femalePop int64
	sqMiles, popPerSqMile   float64
}

func (da *TestSyntheticCountySubdivisionStatsDataAccess) GetPopulation(countyFp string, cousubFp string) int64 {
	return da.pop
}

func (da *TestSyntheticCountySubdivisionStatsDataAccess) GetMalePopulation(countyFp string, cousubFp string) int64 {
	return da.malePop
}

func (da *TestSyntheticCountySubdivisionStatsDataAccess) GetFemalePopulation(countyFp string, cousubFp string) int64 {
	return da.femalePop
}

func (da *TestSyntheticCountySubdivisionStatsDataAccess) GetPopulationPerSquareMile(countyFp string, cousubFp string) float64 {
	return da.popPerSqMile
}

func (da *TestSyntheticCountySubdivisionStatsDataAccess) AddMale(countyFp string, cousubFp string) {
	da.malePop += 1
	da.pop += 1
	da.updatePopPerSqMile()
}

func (da *TestSyntheticCountySubdivisionStatsDataAccess) AddFemale(countyFp string, cousubFp string) {
	da.femalePop += 1
	da.pop += 1
	da.updatePopPerSqMile()
}

func (da *TestSyntheticCountySubdivisionStatsDataAccess) RemoveMale(countyFp string, cousubFp string) {
	da.malePop -= 1
	da.pop -= 1
	da.updatePopPerSqMile()
}

func (da *TestSyntheticCountySubdivisionStatsDataAccess) RemoveFemale(countyFp string, cousubFp string) {
	da.femalePop -= 1
	da.pop -= 1
	da.updatePopPerSqMile()
}

func (da *TestSyntheticCountySubdivisionStatsDataAccess) updatePopPerSqMile() {
	da.popPerSqMile = float64(da.pop) / da.sqMiles
}

// Test harness for PtStatsInterceptor
func TestPtStatsInterceptor(t *testing.T) {

	assert := assert.New(t)

	sqMiles := 100.0
	cousubDA := &TestCountySubdivisionDataAccess{}
	synthCountyStatsDA := &TestSyntheticCountyStatsDataAccess{pop: 20, sqMiles: sqMiles}
	synthCousubStatsDA := &TestSyntheticCountySubdivisionStatsDataAccess{sqMiles: sqMiles}

	s := &PtStatsInterceptor{
		CousubDA:           cousubDA,
		SynthCountyStatsDA: synthCountyStatsDA,
		SynthCousubStatsDA: synthCousubStatsDA,
	}

	body := make([]byte, 512)
	var cousubFp, countyFp string
	var oldCountyPop, oldCountyPopMale, oldCountyPopFemale, oldCousubPop, oldCousubPopMale, oldCousubPopFemale int64
	var newCountyPopPerSqMile, newCousubPopPerSqMile float64

	// 1. Test the interceptor for an AddFemale() operation
	// ====================================================
	body = []byte(`
    {
        "resourceType":"Patient",
        "id":"1295",
        "gender": "female",
        "birthDate":"2009-01-17",
        "address":[
            {
                "line":[
                    "77254 Mafalda Estate",
                    "Apt. 166"
                ],
                "city":"Boston",
                "state":"MA",
                "postalCode":"02163"
            }
        ]
    }
    `)
	cousubFp = s.CousubDA.GetCountySubdivisionFp("Boston")
	countyFp = s.CousubDA.GetCountyFp(cousubFp)

	req, _ := http.NewRequest("POST", "http://example.com/Patient", ioutil.NopCloser(bytes.NewReader(body)))
	c := &gin.Context{Request: req}

	oldCountyPop = s.SynthCountyStatsDA.GetPopulation(countyFp)
	oldCountyPopMale = s.SynthCountyStatsDA.GetMalePopulation(countyFp)
	oldCountyPopFemale = s.SynthCountyStatsDA.GetFemalePopulation(countyFp)

	oldCousubPop = s.SynthCousubStatsDA.GetPopulation(countyFp, cousubFp)
	oldCousubPopMale = s.SynthCousubStatsDA.GetMalePopulation(countyFp, cousubFp)
	oldCousubPopFemale = s.SynthCousubStatsDA.GetFemalePopulation(countyFp, cousubFp)

	s.UpdatePatientStats(c)

	// Test updated county statistics
	assert.Equal(oldCountyPop+1, s.SynthCountyStatsDA.GetPopulation(countyFp), "County population should increment by 1")
	assert.Equal(oldCountyPopMale, s.SynthCountyStatsDA.GetMalePopulation(countyFp), "County male population should not change")
	assert.Equal(oldCountyPopFemale+1, s.SynthCountyStatsDA.GetFemalePopulation(countyFp), "County female population should increment by 1")
	newCountyPopPerSqMile = float64(oldCountyPop+1) / sqMiles
	assert.Equal(newCountyPopPerSqMile, s.SynthCountyStatsDA.GetPopulationPerSquareMile(countyFp), "County population density should change with an increase in population")

	// Test updated cousub statistics
	assert.Equal(oldCousubPop+1, s.SynthCousubStatsDA.GetPopulation(countyFp, cousubFp), "Subdivision population should increment by 1")
	assert.Equal(oldCousubPopMale, s.SynthCousubStatsDA.GetMalePopulation(countyFp, cousubFp), "Subdivision male population should not change")
	assert.Equal(oldCousubPopFemale+1, s.SynthCousubStatsDA.GetFemalePopulation(countyFp, cousubFp), "Subdivision female population should increment by 1")
	newCousubPopPerSqMile = float64(oldCousubPop+1) / sqMiles
	assert.Equal(newCousubPopPerSqMile, s.SynthCousubStatsDA.GetPopulationPerSquareMile(countyFp, cousubFp), "Subdivision population density should change with an increase in population")

	// 2. Test the interceptor for an AddMale() operation
	// ==================================================
	body = []byte(`
    {
        "resourceType":"Patient",
        "id":"1295",
        "gender": "male",
        "birthDate":"2009-01-17",
        "address":[
            {
                "line":[
                    "77254 Mafalda Estate",
                    "Apt. 166"
                ],
                "city":"Bedford",
                "state":"MA",
                "postalCode":"02163"
            }
        ]
    }
    `)
	cousubFp = s.CousubDA.GetCountySubdivisionFp("Bedford")
	countyFp = s.CousubDA.GetCountyFp(cousubFp)

	req, _ = http.NewRequest("POST", "http://example.com/Patient", ioutil.NopCloser(bytes.NewReader(body)))
	c.Request = req

	oldCountyPop = s.SynthCountyStatsDA.GetPopulation(countyFp)
	oldCountyPopMale = s.SynthCountyStatsDA.GetMalePopulation(countyFp)
	oldCountyPopFemale = s.SynthCountyStatsDA.GetFemalePopulation(countyFp)

	oldCousubPop = s.SynthCousubStatsDA.GetPopulation(countyFp, cousubFp)
	oldCousubPopMale = s.SynthCousubStatsDA.GetMalePopulation(countyFp, cousubFp)
	oldCousubPopFemale = s.SynthCousubStatsDA.GetFemalePopulation(countyFp, cousubFp)

	s.UpdatePatientStats(c)

	// Test updated county statistics
	assert.Equal(oldCountyPop+1, s.SynthCountyStatsDA.GetPopulation(countyFp), "County population should increment by 1")
	assert.Equal(oldCountyPopMale+1, s.SynthCountyStatsDA.GetMalePopulation(countyFp), "County male population should increment by 1")
	assert.Equal(oldCountyPopFemale, s.SynthCountyStatsDA.GetFemalePopulation(countyFp), "County female population should not change")
	newCountyPopPerSqMile = float64(oldCountyPop+1) / sqMiles
	assert.Equal(newCountyPopPerSqMile, s.SynthCountyStatsDA.GetPopulationPerSquareMile(countyFp), "County population density should change with an increase in population")

	// Test updated cousub statistics
	assert.Equal(oldCousubPop+1, s.SynthCousubStatsDA.GetPopulation(countyFp, cousubFp), "Subdivision population should increment by 1")
	assert.Equal(oldCousubPopMale+1, s.SynthCousubStatsDA.GetMalePopulation(countyFp, cousubFp), "Subdivision male population should increment by 1")
	assert.Equal(oldCousubPopFemale, s.SynthCousubStatsDA.GetFemalePopulation(countyFp, cousubFp), "Subdivision female population should not change")
	newCousubPopPerSqMile = float64(oldCousubPop+1) / sqMiles
	assert.Equal(newCousubPopPerSqMile, s.SynthCousubStatsDA.GetPopulationPerSquareMile(countyFp, cousubFp), "Subdivision population density should change with an increase in population")

	// 3. Test the interceptor for a RemoveMale() operation
	// ====================================================
	req, _ = http.NewRequest("DELETE", "http://example.com/Patient", ioutil.NopCloser(bytes.NewReader(body)))
	c.Request = req

	oldCountyPop = s.SynthCountyStatsDA.GetPopulation(countyFp)
	oldCountyPopMale = s.SynthCountyStatsDA.GetMalePopulation(countyFp)
	oldCountyPopFemale = s.SynthCountyStatsDA.GetFemalePopulation(countyFp)

	oldCousubPop = s.SynthCousubStatsDA.GetPopulation(countyFp, cousubFp)
	oldCousubPopMale = s.SynthCousubStatsDA.GetMalePopulation(countyFp, cousubFp)
	oldCousubPopFemale = s.SynthCousubStatsDA.GetFemalePopulation(countyFp, cousubFp)

	s.UpdatePatientStats(c)

	// Test updated county statistics
	assert.Equal(oldCountyPop-1, s.SynthCountyStatsDA.GetPopulation(countyFp), "County population should decrement by 1")
	assert.Equal(oldCountyPopMale-1, s.SynthCountyStatsDA.GetMalePopulation(countyFp), "County male population should decrement by 1")
	assert.Equal(oldCountyPopFemale, s.SynthCountyStatsDA.GetFemalePopulation(countyFp), "County female population should not change")
	newCountyPopPerSqMile = float64(oldCountyPop-1) / sqMiles
	assert.Equal(newCountyPopPerSqMile, s.SynthCountyStatsDA.GetPopulationPerSquareMile(countyFp), "County population density should change with an decrease in population")

	// Test updated cousub statistics
	assert.Equal(oldCousubPop-1, s.SynthCousubStatsDA.GetPopulation(countyFp, cousubFp), "Subdivision population should decrement by 1")
	assert.Equal(oldCousubPopMale-1, s.SynthCousubStatsDA.GetMalePopulation(countyFp, cousubFp), "Subdivision male population should decrement by 1")
	assert.Equal(oldCousubPopFemale, s.SynthCousubStatsDA.GetFemalePopulation(countyFp, cousubFp), "Subdivision female population should not change")
	newCousubPopPerSqMile = float64(oldCousubPop-1) / sqMiles
	assert.Equal(newCousubPopPerSqMile, s.SynthCousubStatsDA.GetPopulationPerSquareMile(countyFp, cousubFp), "Subdivision population density should change with an decrease in population")

	// 4. Test the interceptor for a RemoveFemale() operation
	// ======================================================
	body = []byte(`
    {
        "resourceType":"Patient",
        "id":"1295",
        "gender": "female",
        "birthDate":"2009-01-17",
        "address":[
            {
                "line":[
                    "77254 Mafalda Estate",
                    "Apt. 166"
                ],
                "city":"Boston",
                "state":"MA",
                "postalCode":"02163"
            }
        ]
    }
    `)

	cousubFp = s.CousubDA.GetCountySubdivisionFp("Boston")
	countyFp = s.CousubDA.GetCountyFp(cousubFp)

	req, _ = http.NewRequest("DELETE", "http://example.com/Patient", ioutil.NopCloser(bytes.NewReader(body)))
	c.Request = req

	oldCountyPop = s.SynthCountyStatsDA.GetPopulation(countyFp)
	oldCountyPopMale = s.SynthCountyStatsDA.GetMalePopulation(countyFp)
	oldCountyPopFemale = s.SynthCountyStatsDA.GetFemalePopulation(countyFp)

	oldCousubPop = s.SynthCousubStatsDA.GetPopulation(countyFp, cousubFp)
	oldCousubPopMale = s.SynthCousubStatsDA.GetMalePopulation(countyFp, cousubFp)
	oldCousubPopFemale = s.SynthCousubStatsDA.GetFemalePopulation(countyFp, cousubFp)

	s.UpdatePatientStats(c)

	// Test updated county statistics
	assert.Equal(oldCountyPop-1, s.SynthCountyStatsDA.GetPopulation(countyFp), "County population should decrement by 1")
	assert.Equal(oldCountyPopMale, s.SynthCountyStatsDA.GetMalePopulation(countyFp), "County male population should not change")
	assert.Equal(oldCountyPopFemale-1, s.SynthCountyStatsDA.GetFemalePopulation(countyFp), "County female population should decrement by 1")
	newCountyPopPerSqMile = float64(oldCountyPop-1) / sqMiles
	assert.Equal(newCountyPopPerSqMile, s.SynthCountyStatsDA.GetPopulationPerSquareMile(countyFp), "County population density should change with an decrease in population")

	// Test updated cousub statistics
	assert.Equal(oldCousubPop-1, s.SynthCousubStatsDA.GetPopulation(countyFp, cousubFp), "Subdivision population should decrement by 1")
	assert.Equal(oldCousubPopMale, s.SynthCousubStatsDA.GetMalePopulation(countyFp, cousubFp), "Subdivision male population should not change")
	assert.Equal(oldCousubPopFemale-1, s.SynthCousubStatsDA.GetFemalePopulation(countyFp, cousubFp), "Subdivision female population should decrement by 1")
	newCousubPopPerSqMile = float64(oldCousubPop-1) / sqMiles
	assert.Equal(newCousubPopPerSqMile, s.SynthCousubStatsDA.GetPopulationPerSquareMile(countyFp, cousubFp), "Subdivision population density should change with an decrease in population")

	// 5. Test that stats don't change for a body missing patient city
	// ===============================================================
	body = []byte(`
    {
        "resourceType":"Patient",
        "id":"1295",
        "gender": "female",
        "birthDate":"2009-01-17",
        "address":[
            {
                "line":[
                    "77254 Mafalda Estate",
                    "Apt. 166"
                ],
                "state":"MA",
                "postalCode":"02163"
            }
        ]
    }
    `)
	cousubFp = s.CousubDA.GetCountySubdivisionFp("")
	countyFp = s.CousubDA.GetCountyFp(cousubFp)

	req, _ = http.NewRequest("POST", "http://example.com/Patient", ioutil.NopCloser(bytes.NewReader(body)))
	c.Request = req
	testStatsDontChange(t, s, c, countyFp, cousubFp)

	// 6. Test that stats don't change for a body missing patient gender
	// =================================================================
	body = []byte(`
    {
        "resourceType":"Patient",
        "id":"1295",
        "birthDate":"2009-01-17",
        "address":[
            {
                "line":[
                    "77254 Mafalda Estate",
                    "Apt. 166"
                ],
                "city": "Boston",
                "state":"MA",
                "postalCode":"02163"
            }
        ]
    }
    `)
	cousubFp = s.CousubDA.GetCountySubdivisionFp("Boston")
	countyFp = s.CousubDA.GetCountyFp(cousubFp)

	req, _ = http.NewRequest("POST", "http://example.com/Patient", ioutil.NopCloser(bytes.NewReader(body)))
	c.Request = req
	testStatsDontChange(t, s, c, countyFp, cousubFp)

	// 7. Test that stats don't change for HTTP methods that aren't POST or DELETE
	// ===========================================================================
	body = []byte(`
    {
        "resourceType":"Patient",
        "id":"1295",
        "birthDate":"2009-01-17",
        "gender": "male",
        "address":[
            {
                "line":[
                    "77254 Mafalda Estate",
                    "Apt. 166"
                ],
                "city": "Boston",
                "state":"MA",
                "postalCode":"02163"
            }
        ]
    }
    `)
	methods := [6]string{"GET", "PUT", "HEAD", "CONNECT", "OPTIONS", "TRACE"}

	for _, method := range methods {
		req, _ = http.NewRequest(method, "http://example.com/Patient", ioutil.NopCloser(bytes.NewReader(body)))
		c.Request = req
		testStatsDontChange(t, s, c, countyFp, cousubFp)
	}
}

// Reusable testing submodule that verifies no patient statistics have changed
func testStatsDontChange(t *testing.T, s *PtStatsInterceptor, c *gin.Context, countyFp, cousubFp string) {

	oldCountyPop := s.SynthCountyStatsDA.GetPopulation(countyFp)
	oldCountyPopMale := s.SynthCountyStatsDA.GetMalePopulation(countyFp)
	oldCountyPopFemale := s.SynthCountyStatsDA.GetFemalePopulation(countyFp)
	oldCountyPopPerSqMile := s.SynthCountyStatsDA.GetPopulationPerSquareMile(countyFp)

	oldCousubPop := s.SynthCousubStatsDA.GetPopulation(countyFp, cousubFp)
	oldCousubPopMale := s.SynthCousubStatsDA.GetMalePopulation(countyFp, cousubFp)
	oldCousubPopFemale := s.SynthCousubStatsDA.GetFemalePopulation(countyFp, cousubFp)
	oldCousubPopPerSqMile := s.SynthCousubStatsDA.GetPopulationPerSquareMile(countyFp, cousubFp)

	s.UpdatePatientStats(c)

	// Test updated county statistics
	assert.Equal(t, oldCountyPop, s.SynthCountyStatsDA.GetPopulation(countyFp), "County population should not change")
	assert.Equal(t, oldCountyPopMale, s.SynthCountyStatsDA.GetMalePopulation(countyFp), "County male population should not change")
	assert.Equal(t, oldCountyPopFemale, s.SynthCountyStatsDA.GetFemalePopulation(countyFp), "County female population should not change")
	assert.Equal(t, oldCountyPopPerSqMile, s.SynthCountyStatsDA.GetPopulationPerSquareMile(countyFp), "County population density should not change")

	// Test updated cousub statistics
	assert.Equal(t, oldCousubPop, s.SynthCousubStatsDA.GetPopulation(countyFp, cousubFp), "Subdivision population should not change")
	assert.Equal(t, oldCousubPopMale, s.SynthCousubStatsDA.GetMalePopulation(countyFp, cousubFp), "Subdivision male population should not change")
	assert.Equal(t, oldCousubPopFemale, s.SynthCousubStatsDA.GetFemalePopulation(countyFp, cousubFp), "Subdivision female population should not change")
	assert.Equal(t, oldCousubPopPerSqMile, s.SynthCousubStatsDA.GetPopulationPerSquareMile(countyFp, cousubFp), "Subdivision population density should not change")
}
