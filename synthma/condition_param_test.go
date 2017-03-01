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

func TestConditionCodeParamSuite(t *testing.T) {
	suite.Run(t, new(ConditionCodeParamSuite))
}

type ConditionCodeParamSuite struct {
	// see: https://github.com/intervention-engine/ie/blob/master/testutil/MongoSuite.go for original source
	testutil.MongoSuite
}

func (suite *ConditionCodeParamSuite) SetupTest() {
	// Setup the database
	server.Database = suite.DB()
}

func (suite *ConditionCodeParamSuite) TearDownTest() {
	suite.TearDownDB()
}

func (suite *ConditionCodeParamSuite) TearDownSuite() {
	suite.TearDownDBServer()
}

func (suite *ConditionCodeParamSuite) TestConditionCodeParamParserCodeOnly() {
	require := suite.Require()
	assert := suite.Assert()

	p, err := ConditionCodeParamParser(ConditionCodeParamInfo, search.SearchParamData{Value: "60004576"})
	require.NoError(err)
	assert.IsType(new(ConditionCodeParam), p)
	assert.Equal("60004576", p.(*ConditionCodeParam).Code)
}

func (suite *ConditionCodeParamSuite) TestConditionCodeParamParserSystemAndCode() {
	require := suite.Require()
	assert := suite.Assert()

	p, err := ConditionCodeParamParser(ConditionCodeParamInfo, search.SearchParamData{Value: "http://snomed.info/sct|60004576"})
	require.NoError(err)
	assert.IsType(new(ConditionCodeParam), p)
	assert.Equal("http://snomed.info/sct", p.(*ConditionCodeParam).System)
	assert.Equal("60004576", p.(*ConditionCodeParam).Code)
}

func (suite *ConditionCodeParamSuite) TestConditionCodeBSONBuilder() {
	require := suite.Require()
	assert := suite.Assert()

	// Load some conditions into the database
	conditions := make([]models.Condition, 4)
	suite.InsertFixture("conditions", "../fixtures/conditions.json", &conditions)

	// Create the search parameter
	// 62106007 = Concussion, no loss of consciousness
	param, err := ConditionCodeParamParser(ConditionCodeParamInfo, search.SearchParamData{Value: "62106007"})
	require.NoError(err)

	// Run the BSON builder
	obtained, err := ConditionCodeBSONBuilder(param, search.NewMongoSearcher(server.Database, false))
	require.NoError(err)

	// Check the BSON obtained
	expected := bson.M{
		"_id": bson.M{
			"$in": []string{"57ec3d291445d4449de25da2", "57ed3d291445d4449de25da2", "57ef3d291445d4449de25da2"},
		},
	}
	assert.Equal(expected, obtained)
}
