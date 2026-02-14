package service

import (
	"encoding/json"
	"io/ioutil"
	"mallekoppie/ChaosGenerator/internal/master/logic"
	"mallekoppie/ChaosGenerator/internal/master/models"
	"net/http"

	"github.com/Mallekoppie/goslow/platform"
	"go.uber.org/zap"

	"github.com/gorilla/mux"
)

func GetAllAgents(w http.ResponseWriter, r *http.Request) {
	agents, err := logic.GetAllAgents()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(agents)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func DeleteAgent(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		platform.Logger.Error("Unable to read request body: ", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	agent := models.Agent{}

	err = json.Unmarshal(data, &agent)
	if err != nil {
		platform.Logger.Error("Unable to unmarshal request: ", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = logic.DeleteAgent(agent)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func UpdateAgent(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		platform.Logger.Error("Unable to read request body: ", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	agent := models.Agent{}

	err = json.Unmarshal(data, &agent)
	if err != nil {
		platform.Logger.Error("Unable to unmarshal request: ", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = logic.UpdateAgent(agent)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func GetAllTestGroups(w http.ResponseWriter, r *http.Request) {
	testGroups, err := logic.GetAllTestGroups()
	if err != nil {
		platform.Logger.Error("Unable to read test groups from db", zap.Error(err))
	}

	data, err := json.Marshal(testGroups)
	if err != nil {
		platform.Logger.Error("Error while marshalling test groups: ", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func GetTestGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if len(id) < 1 {
		platform.Logger.Error("Must provide valid id when deleting a Test Group")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	group, err := logic.GetTestGroup(id)
	if err != nil && err == platform.ErrNoEntryFoundInDB {
		w.WriteHeader(http.StatusNoContent)
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(group)
	if err != nil {
		platform.Logger.Error("Error marshalling Test Group for get: ", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func AddTestGroup(w http.ResponseWriter, r *http.Request) {
	group := models.TestGroup{}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		platform.Logger.Error("Unable to read body: ", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	platform.Logger.Debug("Request body received: ", zap.String("data", string(data)))

	err = json.Unmarshal(data, &group)
	if err != nil {
		platform.Logger.Error("AddTestGroup: Unable to unmarshal request: ", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = logic.CreateTestGroup(group)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else {
		w.WriteHeader(http.StatusCreated)
		return
	}
}

func UpdateTestGroup(w http.ResponseWriter, r *http.Request) {
	group := models.TestGroup{}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		platform.Logger.Error("Unable to read body: ", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	platform.Logger.Debug("Request body received: ", zap.String("data", string(data)))

	err = json.Unmarshal(data, &group)
	if err != nil {
		platform.Logger.Error("AddTestGroup: Unable to unmarshal request: ", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = logic.UpdateTestGroup(group)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else {
		w.WriteHeader(http.StatusCreated)
		return
	}
}

func DeleteTestGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if len(id) < 1 {
		platform.Logger.Error("Must provide valid id when deleting a Test Group", zap.String("id", id))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err := logic.DeleteTestGroup(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func AddTestCollection(w http.ResponseWriter, r *http.Request) {
	collection := models.TestCollection{}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		platform.Logger.Error("Error reading request body: ", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	err = json.Unmarshal(data, &collection)
	if err != nil {
		platform.Logger.Error("AddTestCollection: Unable to unmarshal request: ", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = logic.AddTestCollection(collection)
	if err != nil && err == logic.ErrParentTestGroupDoesNotExist {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	} else if err != nil {
		platform.Logger.Error("unable to add test collection: ", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func UpdateTestCollection(w http.ResponseWriter, r *http.Request) {
	collection := models.TestCollection{}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		platform.Logger.Error("Error reading request body: ", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	err = json.Unmarshal(data, &collection)
	if err != nil {
		platform.Logger.Error("UpdateTestCollection: Unable to unmarshal request: ", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = logic.UpdateTestCollection(collection)
	if err != nil && err == logic.ErrParentTestGroupDoesNotExist {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	} else if err != nil {
		platform.Logger.Error("unable to add test collection: ", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func DeleteTestCollection(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if len(id) < 1 {
		platform.Logger.Error("Invalid ID sent in path url", zap.String("id", id))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	platform.Logger.Debug("Id Received: ", zap.String("id", id))

	err := logic.DeleteTestCollection(id)
	if err != nil {
		platform.Logger.Error("Unable to delete test collection: ", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func TemplateToCopy(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
