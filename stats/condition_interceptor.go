package stats

import (
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

// NewConditionStatsCreateInterceptor returns an initialized instance of a ConditionStatsCreateInterceptor
func NewConditionStatsCreateInterceptor(pgDataAccess StatsDataAccess, mongoDataAccess server.DataAccessLayer) *ConditionStatsCreateInterceptor {
	return &ConditionStatsCreateInterceptor{PgDataAccess: pgDataAccess, MongoDataAccess: mongoDataAccess}
}

// Before is unused
func (s *ConditionStatsCreateInterceptor) Before(resource interface{}) {}

// After increments disease statistics for given after a new condition resource is created.
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

		if !conditionIsAbated(condition) {
			patient := result.(*models.Patient)
			err = s.PgDataAccess.AddConditionStat(patient, condition)
		}

		if err != nil {
			log.Printf("ConditionStatsCreateInterceptor: %s\n", err.Error())
		}
	}
}

// OnError is unused
func (s *ConditionStatsCreateInterceptor) OnError(err error, resource interface{}) {}

// ConditionStatsUpdateInterceptor intercepts any updates to condition resources in the database
// and updates the Synthetic Mass condition statistics based on the condition's patient's address.
type ConditionStatsUpdateInterceptor struct {
	PgDataAccess    StatsDataAccess
	MongoDataAccess server.DataAccessLayer
	// The state of the condition before the database update, for comparison after the database update
	conditionsBefore map[string]*models.Condition
}

// NewConditionStatsUpdateInterceptor returns an initialized instance of ConditionStatsUpdateInterceptor
func NewConditionStatsUpdateInterceptor(pgDataAccess StatsDataAccess, mongoDataAccess server.DataAccessLayer) *ConditionStatsUpdateInterceptor {
	interceptor := &ConditionStatsUpdateInterceptor{PgDataAccess: pgDataAccess, MongoDataAccess: mongoDataAccess}
	interceptor.conditionsBefore = make(map[string]*models.Condition)
	return interceptor
}

// Before caches a condition resource before it's updated, for comparison after the update.
func (s *ConditionStatsUpdateInterceptor) Before(resource interface{}) {
	condition, ok := resource.(*models.Condition)

	if ok {
		s.conditionsBefore[condition.Id] = condition
	} else {
		log.Println("ConditionStatsUpdateInterceptor:Before: Failed to cache condition before update")
	}
}

// After compares the updated condition resource to the cached condition resource (from before the update), then
// updates population statistics based on that resource's patient's address.
func (s *ConditionStatsUpdateInterceptor) After(resource interface{}) {
	newCondition, ok := resource.(*models.Condition)

	if ok {

		oldCondition, found := s.conditionsBefore[newCondition.Id]

		if !found {
			log.Println("ConditionStatsUpdateInterceptor: After: Could not find cached condition")
			return
		}

		// free up memory
		delete(s.conditionsBefore, oldCondition.Id)

		// see if the condition is now abated, so we can stop tracking stats for it
		if !conditionIsAbated(oldCondition) && conditionIsAbated(newCondition) {

			if newCondition.Subject == nil {
				log.Printf("ConditionStatsUpdateInterceptor: Condition %s has no subject\n", newCondition.Id)
				return
			}

			result, err := s.MongoDataAccess.Get(newCondition.Subject.ReferencedID, "Patient")

			if err != nil {
				log.Printf("ConditionStatsUpdateInterceptor: Failed to get patient for condition %s\n", newCondition.Id)
				return
			}

			patient := result.(*models.Patient)
			err = s.PgDataAccess.RemoveConditionStat(patient, newCondition)

			if err != nil {
				log.Printf("ConditionStatsUpdateInterceptor: %s\n", err.Error())
			}
		}
	}
}

// OnError is unused
func (s *ConditionStatsUpdateInterceptor) OnError(err error, resource interface{}) {}

// ConditionStatsDeleteInterceptor intercepts any deleted condition resources
// and updates the Synthetic Mass condition statistics based on the condition's patient's address.
type ConditionStatsDeleteInterceptor struct {
	PgDataAccess    StatsDataAccess
	MongoDataAccess server.DataAccessLayer
}

// NewConditionStatsDeleteInterceptor returns an initialized instance of ConditionStatsDeleteInterceptor
func NewConditionStatsDeleteInterceptor(pgDataAccess StatsDataAccess, mongoDataAccess server.DataAccessLayer) *ConditionStatsDeleteInterceptor {
	return &ConditionStatsDeleteInterceptor{PgDataAccess: pgDataAccess, MongoDataAccess: mongoDataAccess}
}

// Before is unused
func (s *ConditionStatsDeleteInterceptor) Before(resource interface{}) {}

// After decrements disease statistics for a given condition resource after it's deleted.
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

		if !conditionIsAbated(condition) {
			patient := result.(*models.Patient)
			err = s.PgDataAccess.RemoveConditionStat(patient, condition)
		}

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
