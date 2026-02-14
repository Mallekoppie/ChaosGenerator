package repositories

import (
	"mallekoppie/ChaosGenerator/internal/master/models"

	"github.com/Mallekoppie/goslow/platform"
	"go.uber.org/zap"

	"encoding/json"
	"errors"
)

const (
	bucketTestGroup      string = "testgroup"
	bucketTestCollection string = "testcollection"
	bucketAgent          string = "agent"
)

var (
	ErrUpdateCountWrong                 = errors.New("Incorrect update count")
	ErrNoGroupExistsForTestCollectionId = errors.New("No group exists for ID")
)

func init() {

}

func AddTestGroup(testGroup models.TestGroup) error {
	testCollection := testGroup.TestCollections
	testGroup.TestCollections = nil
	err := platform.Database.BoltDb.SaveObject(bucketTestGroup, testGroup.ID, testGroup)
	if err != nil {
		platform.Logger.Error("Error saving TestGroup to DB", zap.Error(err))
		return err
	}

	for _, v := range testCollection {
		err = platform.Database.BoltDb.SaveObject(bucketTestCollection, v.ID, v)
		if err != nil {
			platform.Logger.Error("Error saving TestCollection", zap.Error(err))
			return err
		}
	}

	return nil
}

func GetTestGroup(id string) (testGroup models.TestGroup, err error) {

	err = platform.Database.BoltDb.ReadObject(bucketTestGroup, id, &testGroup)
	if err != nil {
		platform.Logger.Error("Error reading testgroup from database", zap.String("id", id))
		return
	}

	collections, err := platform.Database.BoltDb.ReadAllObjects(bucketTestCollection)
	if err != nil {
		platform.Logger.Error("Error reading Test Collections for Test Group", zap.String("test_group_id", id), zap.Error(err))
		return
	}

	testCollections := make([]models.TestCollection, 0)

	for _, v := range collections {
		col := models.TestCollection{}
		err = json.Unmarshal([]byte(v), &col)
		if err != nil {
			platform.Logger.Error("Error unmarshalling test collection json data", zap.Error(err))
			return
		}

		if col.GroupId == id {
			testCollections = append(testCollections, col)
		}
	}

	testGroup.TestCollections = testCollections

	return testGroup, nil
}

func DoesTestGroupExist(id string) (bool, error) {
	testGroup := models.TestGroup{}
	err := platform.Database.BoltDb.ReadObject(bucketTestGroup, id, &testGroup)
	if err != nil && err == platform.ErrNoEntryFoundInDB {
		return false, nil
	} else if err != nil {
		platform.Logger.Error("Error reading object to see if it exists", zap.Error(err))
		return false, err
	}

	return true, nil
}

func DeleteAllTestGroups() error {

	err := platform.Database.BoltDb.RemoveBucket(bucketTestCollection)
	if err != nil {
		platform.Logger.Error("Error remove test collections", zap.Error(err))
		return err
	}
	err = platform.Database.BoltDb.RemoveBucket(bucketTestGroup)
	if err != nil {
		platform.Logger.Error("Error remove test groups", zap.Error(err))
		return err
	}

	return nil
}

func DeleteTestGroup(id string) error {

	result, err := platform.Database.BoltDb.ReadAllObjects(bucketTestCollection)
	if err != nil {
		platform.Logger.Error("Error reading test collections before removing them", zap.Error(err))
		return err
	}

	for _, v := range result {
		col := models.TestCollection{}
		err = json.Unmarshal([]byte(v), &col)
		if err != nil {
			platform.Logger.Error("Error unmarshalling test collection", zap.Error(err))
			return err
		}

		if col.GroupId == id {
			platform.Logger.Debug("Removing test collection as part of test group deletion", zap.String("id", col.ID))
			platform.Database.BoltDb.RemoveObject(bucketTestCollection, col.ID)
		}
	}

	err = platform.Database.BoltDb.RemoveObject(bucketTestGroup, id)
	if err != nil {
		platform.Logger.Error("Error removing test group", zap.Error(err))
		return err
	}

	return nil
}

func DeleteTestCollection(id string) error {
	err := platform.Database.BoltDb.RemoveObject(bucketTestCollection, id)
	if err != nil {
		platform.Logger.Error("Error removing test collection", zap.Error(err))
		return err
	}

	return nil
}

func UpdateTestGroup(testGroup models.TestGroup) error {

	err := platform.Database.BoltDb.SaveObject(bucketTestGroup, testGroup.ID, testGroup)
	if err != nil {
		platform.Logger.Error("Error saving Test Group", zap.Error(err))
		return err
	}

	return nil
}

func UpdateTestCollection(col models.TestCollection) error {

	err := platform.Database.BoltDb.SaveObject(bucketTestCollection, col.ID, col)
	if err != nil {
		platform.Logger.Error("Error saving test collection", zap.Error(err), zap.String("id", col.ID))
		return err
	}

	return nil
}

func GetAllTestGroups() (testGroups []models.TestGroup, err error) {

	result, err := platform.Database.BoltDb.ReadAllObjects(bucketTestGroup)
	if err != nil {
		platform.Logger.Error("Error reading all test groups from DB", zap.Error(err))
		return nil, err
	}

	testGroups = make([]models.TestGroup, 0)

	for _, v := range result {
		group := models.TestGroup{}
		err = json.Unmarshal([]byte(v), &group)
		if err != nil {
			platform.Logger.Error("Error unmarchalling TestGroup", zap.Error(err))
			return nil, err
		}

		testGroups = append(testGroups, group)
	}

	return testGroups, nil
}

func AddTestCollection(tests models.TestCollection) error {

	err := platform.Database.BoltDb.SaveObject(bucketTestCollection, tests.ID, tests)
	if err != nil {
		platform.Logger.Error("Error saving test collection", zap.Error(err))
		return err
	}

	return nil
}

func GetTestCollectionsForGroup(id string) (tests []models.TestCollection, err error) {

	result, err := platform.Database.BoltDb.ReadAllObjects(bucketTestCollection)
	if err != nil {
		platform.Logger.Error("Error reading all test collections", zap.Error(err))
		return nil, err
	}

	tests = make([]models.TestCollection, 0)

	for _, v := range result {
		col := models.TestCollection{}
		err = json.Unmarshal([]byte(v), &col)
		if err != nil {
			platform.Logger.Error("Error unmarshalling test collection", zap.Error(err))
			return nil, err
		}

		if col.GroupId == id {
			tests = append(tests, col)
		}
	}

	return tests, nil
}

func AddAgent(agent models.Agent) error {
	err := platform.Database.BoltDb.SaveObject(bucketAgent, agent.Id, agent)
	if err != nil {
		platform.Logger.Error("Error saving agent", zap.Error(err))
		return err
	}

	return nil
}

func DeleteAgent(id string) error {
	err := platform.Database.BoltDb.RemoveObject(bucketAgent, id)
	if err != nil {
		platform.Logger.Error("Error removing agent", zap.Error(err))
		return err
	}

	return nil
}

func DeleteAllAgents() error {
	err := platform.Database.BoltDb.RemoveBucket(bucketAgent)
	if err != nil {
		platform.Logger.Error("Error removing all agents", zap.Error(err))
		return err
	}

	return nil
}

func UpdateAgent(agent models.Agent) error {
	err := platform.Database.BoltDb.SaveObject(bucketAgent, agent.Id, agent)
	if err != nil {
		platform.Logger.Error("Error Updating agent", zap.Error(err))
		return err
	}

	return nil
}

func GetAllAgents() (agents []models.Agent, err error) {
	agents = make([]models.Agent, 0)

	results, err := platform.Database.BoltDb.ReadAllObjects(bucketAgent)
	if err != nil && err == platform.ErrNoEntryFoundInDB {
		return agents, nil
	}
	if err != nil {
		platform.Logger.Error("Error retrieving agents", zap.Error(err))
		return nil, err
	}

	for _, v := range results {
		agent := models.Agent{}
		err = json.Unmarshal([]byte(v), &agent)
		if err != nil {
			platform.Logger.Error("Error unmarshalling agent", zap.Error(err))
			return nil, err
		}

		agents = append(agents, agent)
	}

	return agents, nil
}
