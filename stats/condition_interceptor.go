package stats

import (
	"errors"
	"log"

	"github.com/intervention-engine/fhir/models"
	"github.com/intervention-engine/fhir/server"
)

// ConditionStatsCreateInterceptor intercepts any new condition resources added to the database
// and updates the Synthetic Mass condition statistics based on the condition's patient's address.
type ConditionStatsCreateInterceptor struct {
	PgDataAccess    StatsDataAccess
	MongoDataAccess server.DataAccessLayer
}

// Before is unused
func (s *ConditionStatsCreateInterceptor) Before(resource interface{}) {}

func (s *ConditionStatsCreateInterceptor) After(resource interface{}) {
	condition, ok := resource.(*models.Condition)

	if ok {

		var err error

		if condition.Subject == nil {
			log.Printf("ConditionStatsCreateInterceptor: Condition %s has no subject\n", condition.Id)
			return
		}

		result, err := s.MongoDataAccess.Get(condition.Subject.ReferencedID, "Patient")

		if err != nil {
			log.Printf("ConditionStatsCreateInterceptor: Failed to get patient for condition %s\n", condition.Id)
			return
		}

		patient := result.(*models.Patient)
		err = s.PgDataAccess.AddConditionStat(patient, condition)

		if err != nil {
			log.Printf("ConditionStatsCreateInterceptor: %s\n", err.Error())
		}
	}
}

// OnError is unused
func (s *ConditionStatsCreateInterceptor) OnError(err error, resource interface{}) {}

// ConditionStatsCreateInterceptor intercepts any updates to condition resources in the database
// and updates the Synthetic Mass condition statistics based on the condition's patient's address.
type ConditionStatsUpdateInterceptor struct {
	PgDataAccess    StatsDataAccess
	MongoDataAccess server.DataAccessLayer
	// The state of the condition before the database update, for comparison after the database update
	conditionBefore *models.Condition
	// Tracks if the interceptor failed to cache the condition model before the update
	cacheError error
}

func (s *ConditionStatsUpdateInterceptor) Before(resource interface{}) {
	condition, ok := resource.(*models.Condition)

	if ok {
		s.conditionBefore = condition
	} else {
		errmsg := "ConditionStatsUpdateInterceptor:Before: Failed to cache condition before update"
		s.cacheError = errors.New(errmsg)
		log.Println(errmsg)
	}
}

func (s *ConditionStatsUpdateInterceptor) After(resource interface{}) {
	condition, ok := resource.(*models.Condition)

	if ok && s.cacheError == nil {
		// see if the condition is now abated, so we can stop tracking stats for it
		if !conditionIsAbated(s.conditionBefore) && conditionIsAbated(condition) {

			if condition.Subject == nil {
				log.Printf("ConditionStatsUpdateInterceptor: Condition %s has no subject\n", condition.Id)
				return
			}

			result, err := s.MongoDataAccess.Get(condition.Subject.ReferencedID, "Patient")

			if err != nil {
				log.Printf("ConditionStatsUpdateInterceptor: Failed to get patient for condition %s\n", condition.Id)
				return
			}

			patient := result.(*models.Patient)
			err = s.PgDataAccess.RemoveConditionStat(patient, condition)

			if err != nil {
				log.Printf("ConditionStatsUpdateInterceptor: %s\n", err.Error())
			}
		}
	}
}

// OnError is unused
func (s *ConditionStatsUpdateInterceptor) OnError(err error, resource interface{}) {}

// ConditionStatsCreateInterceptor intercepts any deleted condition resources
// and updates the Synthetic Mass condition statistics based on the condition's patient's address.
type ConditionStatsDeleteInterceptor struct {
	PgDataAccess    StatsDataAccess
	MongoDataAccess server.DataAccessLayer
}

// Before is unused
func (s *ConditionStatsDeleteInterceptor) Before(resource interface{}) {}

func (s *ConditionStatsDeleteInterceptor) After(resource interface{}) {
	condition, ok := resource.(*models.Condition)

	if ok {

		var err error

		if condition.Subject == nil {
			log.Printf("ConditionStatsDeleteInterceptor: Condition %s has no subject\n", condition.Id)
			return
		}

		result, err := s.MongoDataAccess.Get(condition.Subject.ReferencedID, "Patient")

		if err != nil {
			log.Printf("ConditionStatsDeleteInterceptor: Failed to get patient for condition %s\n", condition.Id)
			return
		}

		patient := result.(*models.Patient)
		err = s.PgDataAccess.RemoveConditionStat(patient, condition)

		if err != nil {
			log.Printf("ConditionStatsDeleteInterceptor: %s\n", err.Error())
		}
	}
}

// OnError is unused
func (s *ConditionStatsDeleteInterceptor) OnError(err error, resource interface{}) {}

func conditionIsAbated(condition *models.Condition) bool {
	if condition.AbatementDateTime != nil ||
		condition.AbatementAge != nil ||
		condition.AbatementBoolean != nil ||
		condition.AbatementPeriod != nil ||
		condition.AbatementRange != nil ||
		condition.AbatementString != "" {
		return true
	}
	return false
}
