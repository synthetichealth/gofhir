package synthma

import (
	"testing"

	"gopkg.in/mgo.v2/bson"

	"github.com/intervention-engine/fhir/models"
	"github.com/intervention-engine/fhir/search"
	"github.com/intervention-engine/fhir/server"
	"github.com/stretchr/testify/suite"
	"github.com/synthetichealth/gofhir/testutil"
)

func TestUSCoreParamsSuite(t *testing.T) {
	suite.Run(t, new(USCoreParamsSuite))
}

type USCoreParamsSuite struct {
	// see: https://github.com/intervention-engine/ie/blob/master/testutil/MongoSuite.go for original source
	testutil.MongoSuite
}

func (suite *USCoreParamsSuite) SetupTest() {
	// Setup the database
	server.Database = suite.DB()
}

func (suite *USCoreParamsSuite) TearDownTest() {
	suite.TearDownDB()
}

func (suite *USCoreParamsSuite) TearDownSuite() {
	suite.TearDownDBServer()
}

func (suite *USCoreParamsSuite) TestUSCoreRaceParamParserCodeOnly() {
	require := suite.Require()
	assert := suite.Assert()

	p, err := USCoreRaceParamParser(USCoreRaceParamInfo, search.SearchParamData{Value: "2106-3"})
	require.NoError(err)
	assert.IsType(new(USCoreRaceParam), p)
	assert.Equal("2106-3", p.(*USCoreRaceParam).Code)
}

func (suite *USCoreParamsSuite) TestUSCoreRaceParamParserSystemAndCode() {
	require := suite.Require()
	assert := suite.Assert()

	p, err := USCoreRaceParamParser(USCoreRaceParamInfo, search.SearchParamData{Value: "http://hl7.org/fhir/v3/Race|2106-3"})
	require.NoError(err)
	assert.IsType(new(USCoreRaceParam), p)
	assert.Equal("http://hl7.org/fhir/v3/Race", p.(*USCoreRaceParam).System)
	assert.Equal("2106-3", p.(*USCoreRaceParam).Code)
}

func (suite *USCoreParamsSuite) TestUSCoreRaceBSONBuilder() {
	require := suite.Require()
	assert := suite.Assert()

	// Load some patients into the database
	patients := make([]models.Patient, 2)
	suite.InsertFixture("patients", "../fixtures/patients.json", &patients)

	// Create the search parameter
	// 2106-3 = White
	param, err := USCoreRaceParamParser(USCoreRaceParamInfo, search.SearchParamData{Value: "2106-3"})
	require.NoError(err)

	// Run the BSON builder
	obtained, err := USCoreRaceBSONBuilder(param, search.NewMongoSearcher(server.Database, false))
	require.NoError(err)

	// Check the BSON obtained
	expected := bson.M{
		"extension.us-core-race.coding.system": "http://hl7.org/fhir/v3/Race",
		"extension.us-core-race.coding.code":   "2106-3",
	}
	assert.Equal(expected, obtained)
}

func (suite *USCoreParamsSuite) TestUSCoreEthnicityParamParserCodeOnly() {
	require := suite.Require()
	assert := suite.Assert()

	p, err := USCoreEthnicityParamParser(USCoreEthnicityParamInfo, search.SearchParamData{Value: "2186-5"})
	require.NoError(err)
	assert.IsType(new(USCoreEthnicityParam), p)
	assert.Equal("2186-5", p.(*USCoreEthnicityParam).Code)
}

func (suite *USCoreParamsSuite) TestUSCoreEthnicityParamParserSystemAndCode() {
	require := suite.Require()
	assert := suite.Assert()

	p, err := USCoreEthnicityParamParser(USCoreEthnicityParamInfo, search.SearchParamData{Value: "http://hl7.org/fhir/v3/Ethnicity|2186-5"})
	require.NoError(err)
	assert.IsType(new(USCoreEthnicityParam), p)
	assert.Equal("http://hl7.org/fhir/v3/Ethnicity", p.(*USCoreEthnicityParam).System)
	assert.Equal("2186-5", p.(*USCoreEthnicityParam).Code)
}

func (suite *USCoreParamsSuite) TestUSCoreEthnicityBSONBuilder() {
	require := suite.Require()
	assert := suite.Assert()

	// Load some patient into the database
	patients := make([]models.Patient, 2)
	suite.InsertFixture("patients", "../fixtures/patients.json", &patients)

	// Create the search parameter
	// 2186-5 = Nonhispanic
	param, err := USCoreEthnicityParamParser(USCoreEthnicityParamInfo, search.SearchParamData{Value: "2186-5"})
	require.NoError(err)

	// Run the BSON builder
	obtained, err := USCoreEthnicityBSONBuilder(param, search.NewMongoSearcher(server.Database, false))
	require.NoError(err)

	// Check the BSON obtained
	expected := bson.M{
		"extension.us-core-ethnicity.coding.system": "http://hl7.org/fhir/v3/Ethnicity",
		"extension.us-core-ethnicity.coding.code":   "2186-5",
	}
	assert.Equal(expected, obtained)
}
