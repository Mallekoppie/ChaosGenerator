package main

import (
	"mallekoppie/ChaosGenerator/internal/master/service"
	"net/http"

	"github.com/Mallekoppie/goslow/platform"
)

const (
	allowedContentType       string = "application/json"
	allowedContentTypeDelete string = "application/json;charset=UTF-8"
)

var Routes = platform.Routes{
	platform.Route{
		Path:          "/testgroups/{id}",
		Method:        http.MethodGet,
		HandlerFunc:   service.GetTestGroup,
		SlaMs:         100,
		RolesRequired: []string{"user"},
	},
	// Only during development
	platform.Route{
		Path:          "/testgroups/{id}",
		Method:        http.MethodOptions,
		HandlerFunc:   service.TemplateToCopy,
		SlaMs:         100,
		RolesRequired: []string{"user"},
	},
	platform.Route{
		Path:          "/testgroups",
		Method:        http.MethodGet,
		HandlerFunc:   service.GetAllTestGroups,
		SlaMs:         100,
		RolesRequired: []string{"user"},
	},
	platform.Route{
		Path:               "/testgroups",
		Method:             http.MethodPost,
		HandlerFunc:        service.AddTestGroup,
		SlaMs:              100,
		RolesRequired:      []string{"user"},
		AllowedContentType: allowedContentType,
	},
	platform.Route{
		Path:               "/testgroups",
		Method:             http.MethodPut,
		HandlerFunc:        service.UpdateTestGroup,
		SlaMs:              100,
		RolesRequired:      []string{"user"},
		AllowedContentType: allowedContentType,
	},
	platform.Route{
		Path:          "/testgroups/{id}",
		Method:        http.MethodDelete,
		HandlerFunc:   service.DeleteTestGroup,
		SlaMs:         100,
		RolesRequired: []string{"user"},
	},
	// Only during development
	platform.Route{
		Path:          "/testgroups",
		Method:        http.MethodOptions,
		HandlerFunc:   service.TemplateToCopy,
		SlaMs:         100,
		RolesRequired: []string{"user"},
	},
	// Only during development
	platform.Route{
		Path:          "/testgroups/{id}",
		Method:        http.MethodOptions,
		HandlerFunc:   service.TemplateToCopy,
		SlaMs:         100,
		RolesRequired: []string{"user"},
	},
	platform.Route{
		Path:               "/testcollections",
		Method:             http.MethodPost,
		HandlerFunc:        service.AddTestCollection,
		SlaMs:              100,
		RolesRequired:      []string{"user"},
		AllowedContentType: allowedContentType,
	},
	platform.Route{
		Path:               "/testcollections",
		Method:             http.MethodPut,
		HandlerFunc:        service.UpdateTestCollection,
		SlaMs:              100,
		RolesRequired:      []string{"user"},
		AllowedContentType: allowedContentType,
	},
	// Only during development
	platform.Route{
		Path:          "/testcollections",
		Method:        http.MethodOptions,
		HandlerFunc:   service.TemplateToCopy,
		SlaMs:         100,
		RolesRequired: []string{"user"},
	},
	platform.Route{
		Path:          "/testcollections/{id}",
		Method:        http.MethodDelete,
		HandlerFunc:   service.DeleteTestCollection,
		SlaMs:         100,
		RolesRequired: []string{"user"},
	},
	// Only during development
	platform.Route{
		Path:          "/testcollections/{id}",
		Method:        http.MethodOptions,
		HandlerFunc:   service.TemplateToCopy,
		SlaMs:         100,
		RolesRequired: []string{"user"},
	},
	platform.Route{
		Path:          "/agents",
		Method:        http.MethodGet,
		HandlerFunc:   service.GetAllAgents,
		SlaMs:         100,
		RolesRequired: []string{"user"},
	},
	platform.Route{
		Path:               "/agents",
		Method:             http.MethodPut,
		HandlerFunc:        service.UpdateAgent,
		SlaMs:              100,
		RolesRequired:      []string{"user"},
		AllowedContentType: allowedContentType,
	},
	platform.Route{
		Path:               "/agents",
		Method:             http.MethodDelete,
		HandlerFunc:        service.DeleteAgent,
		SlaMs:              100,
		RolesRequired:      []string{"user"},
		AllowedContentType: allowedContentTypeDelete,
	},
	// Only during development
	platform.Route{
		Path:          "/agents",
		Method:        http.MethodOptions,
		HandlerFunc:   service.TemplateToCopy,
		SlaMs:         100,
		RolesRequired: []string{"user"},
	},
}
