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
    //"encoding/json"

    "github.com/gin-gonic/gin"

    "github.com/jinzhu/gorm"
    _ "github.com/jinzhu/gorm/dialects/postgres"
)

// Middleware that handles the interceptor
type PtStatsInterceptor struct {
    DB *gorm.DB
}

/*
GORM models that map to Postgres synth_ma tables
*/
type CountySubdivision struct {

    StateFp string `gorm:"column:statefp"`
    CountyFp string `gorm:"column:countyfp"`
    CountySubdivisonFp string `gorm:"column:cousubfp"`
    Name string `gorm:"column:name"`
}

// Set CountySubdivision table name to be "tiger.cousub"
func (CountySubdivision) TableName() string {
    return "tiger.cousub"
}

type SyntheticCountyStatistics struct {

    CountyName  string `gorm:"column:ct_name"`
    CountyFp    string `gorm:"column:ct_fips"`
    SquareMileage   float64 `gorm:"column:sq_mi"`
    Population  int64   `gorm:"column:pop"`
    PopulationMale  int64   `gorm:"column:pop_male"`
    PopulationFemale    int64   `gorm:"column:pop_female"`
    PopulationPerSquareMile float64 `gorm:"column:pop_sm"`
}

// Set SyntheticCountyStatistics table name to be "synth_ma.synth_county_stats"
func (SyntheticCountyStatistics) TableName() string {
    return "synth_ma.synth_county_stats"
}

type SyntheticCountySubdivisionStatistics struct {

    CountyName  string `gorm:"column:ct_name"`
    CountyFp    string `gorm:"column:ct_fips"`
    CountySubdivisionName   string `gorm:"column:"cs_name"`
    SquareMileage   float64 `gorm:"column:sq_mi"`
    Population  int64   `gorm:"column:pop"`
    PopulationMale  int64   `gorm:"column:pop_male"`
    PopulationFemale    int64   `gorm:"column:pop_female"`
    PopulationPerSquareMile float64 `gorm:"column:pop_sm"`
}

// Set SyntheticCountrySubdivisionStatistics table name to "synth_ma.synth_cousub_stats"
func (SyntheticCountySubdivisionStatistics) TableName() string {
    return "synth_ma.synth_cousub_stats"
}

// Handler is registered with the GoFHIR server and invoked on every request
func (s *PtStatsInterceptor) Handler(c *gin.Context) {

    // Only handle Create, Update, and Delete operations for a Patient
    if c.Request != nil && 
       c.Request.URL.Path == "/Patient" &&
       (c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "DELETE") {

        // Read the body and close the stream
        body, _ := ioutil.ReadAll(c.Request.Body)
        c.Request.Body.Close()

        // Do something with the patient JSON body


        // We need to replenish the body since we drained the stream
        c.Request.Body = ioutil.NopCloser(bytes.NewReader(body))
    }

    // Go to the next handler
    c.Next()
}



