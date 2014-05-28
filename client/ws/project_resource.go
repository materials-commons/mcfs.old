package ws

import (
	"fmt"
	"github.com/emicklei/go-restful"
	"github.com/materials-commons/mcfs/client"
	"net/http"
)

type projectResource struct {
	*materials.ProjectDB
}

func newProjectResource(container *restful.Container) error {
	p := materials.CurrentUserProjectDB()

	projectResource := projectResource{
		ProjectDB: p,
	}
	projectResource.register(container)

	return nil
}

func (p projectResource) register(container *restful.Container) {
	ws := new(restful.WebService)
	ws.Path("/projects").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON, restful.MIME_XML)

	ws.Route(ws.GET("").Filter(JSONPFilter).To(p.allProjects).
		Doc("list all projects").
		Writes([]materials.Project{}))

	ws.Route(ws.GET("/{project-name}").Filter(JSONPFilter).To(p.getProject).
		Doc("Retrieve a particular project").
		Param(ws.PathParameter("project-name", "name of the project").DataType("string")).
		Writes(materials.Project{}))

	ws.Route(ws.GET("/{project-name}/tree").Filter(JSONPFilter).To(p.getProjectTree).
		Doc("Retrieve the directory/file tree for the project").
		Param(ws.PathParameter("original-project-name", "original name of the project").
		DataType("string")))

	ws.Route(ws.POST("").To(p.newProject).
		Doc("Create a new project").
		Reads(materials.Project{}))

	ws.Route(ws.PUT("/{project-name}").To(p.updateProject).
		Doc("Updates the project").
		Param(ws.PathParameter("project-name", "name of the project").DataType("string")).
		Reads(materials.Project{}).
		Writes(materials.Project{}))

	ws.Route(ws.GET("/{project-name}/changes").To(p.getProjectChanges).
		Doc("Lists all the file system changes for a project").
		Param(ws.PathParameter("project-name", "name of the project").DataType("string")).
		Writes([]materials.ProjectFileChange{}))

	ws.Route(ws.PUT("/{project-name}/track").To(p.updateTracking).
		Doc("Updates the file tracking for the project").
		Reads(materials.TrackingOptions{}).
		Param(ws.PathParameter("project-name", "name of the project").DataType("string")))

	container.Add(ws)
}

func (p projectResource) allProjects(request *restful.Request, response *restful.Response) {
	if len(p.Projects()) == 0 {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusNotFound, "User has no projects.")
	} else {
		response.WriteEntity(p.Projects())
	}
}

func (p projectResource) getProject(request *restful.Request, response *restful.Response) {
	projectName := request.PathParameter("project-name")
	project, found := p.Find(projectName)
	if found {
		response.WriteEntity(project)
	} else {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusNotFound,
			fmt.Sprintf("Project not found: %s", projectName))
	}
}

func (p projectResource) getProjectTree(request *restful.Request, response *restful.Response) {
	projectName := request.PathParameter("project-name")

	if project, found := p.Find(projectName); !found {
		response.WriteErrorString(http.StatusNotFound,
			fmt.Sprintf("Project not found: %s", projectName))
	} else if tree, err := project.Tree(); err != nil {
		response.WriteErrorString(http.StatusNotAcceptable, err.Error())
	} else {
		response.WriteEntity(tree)
	}
}

func (p *projectResource) newProject(request *restful.Request, response *restful.Response) {
	project := new(materials.Project)
	err := request.ReadEntity(&project)
	if err != nil {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusNotAcceptable, err.Error())
		return
	}

	err = p.Add(*project)
	if err != nil {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusNotAcceptable, err.Error())
		return
	}

	response.WriteHeader(http.StatusCreated)
	response.WriteEntity(project)
}

func (p *projectResource) updateProject(request *restful.Request, response *restful.Response) {
	originalProjectName := request.PathParameter("original-project-name")
	project := new(materials.Project)
	err := request.ReadEntity(&project)
	if err != nil {
		response.WriteErrorString(http.StatusNotAcceptable, err.Error())
		return
	}

	originalProject, found := p.Find(originalProjectName)
	if !found {
		response.WriteErrorString(http.StatusNotFound,
			fmt.Sprintf("Project not found '%s'", originalProjectName))
		return
	}

	if project.Name != originalProjectName {
		p.Remove(originalProjectName)
		project.Status = originalProject.Status
		err = p.Add(*project)
	} else {
		err = p.Update(func() *materials.Project {
			originalProject.Name = project.Name
			originalProject.Path = project.Path
			return originalProject
		})
	}

	if err != nil {
		response.WriteErrorString(http.StatusNotAcceptable, err.Error())
	} else {
		response.WriteEntity(project)
	}
}

func (p *projectResource) getProjectChanges(request *restful.Request, response *restful.Response) {
	p.performProjectOperation(request, response, func(project *materials.Project) {
		changes := []materials.ProjectFileChange{}
		for _, change := range project.Changes {
			changes = append(changes, change)
		}
		response.WriteEntity(changes)
	})
}

func (p *projectResource) updateTracking(request *restful.Request, response *restful.Response) {
	trackingOptions := materials.TrackingOptions{}
	request.ReadEntity(&trackingOptions)

	p.performProjectOperation(request, response, func(project *materials.Project) {
		go func() {
			if err := project.Walk(&trackingOptions); err != nil {
				msg := fmt.Sprintf("Error updating tracking for project %s: %s\n", project.Name, err.Error())
				response.WriteErrorString(http.StatusInternalServerError, msg)
			}
		}()

		msg := fmt.Sprintf("Updating tracking for project %s\n", project.Name)
		response.WriteErrorString(http.StatusOK, msg)
	})
}

func (p *projectResource) performProjectOperation(request *restful.Request, response *restful.Response, f func(project *materials.Project)) {
	projectName := request.PathParameter("project-name")
	if project, found := p.Find(projectName); found {
		f(project)
	} else {
		response.WriteErrorString(http.StatusNotFound,
			fmt.Sprintf("Project not found: %s\n", projectName))
	}
}
