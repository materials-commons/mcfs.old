go-restful
==========

package for building REST-style Web Services using Google Go

REST asks developers to use HTTP methods explicitly and in a way that's consistent with the protocol definition. This basic REST design principle establishes a one-to-one mapping between create, read, update, and delete (CRUD) operations and HTTP methods. According to this mapping:

- GET = Retrieve a representation of a resource
- POST = Create if you are sending content to the server to create a subordinate of the specified resource collection, using some server-side algorithm.
- PUT = Create iff you are sending the full content of the specified resource (URI).
- PUT = Update iff you are updating the full content of the specified resource.
- DELETE = Delete if you are requesting the server to delete the resource
- PATCH = Update partial content of a resource
- OPTIONS = Get information about the communication options for the request URI
    
### Example

	ws := new(restful.WebService)
	ws.
		Path("/users").
		Consumes(restful.MIME_XML, restful.MIME_JSON).
		Produces(restful.MIME_JSON, restful.MIME_XML)

	ws.Route(ws.GET("/{user-id}").To(u.findUser).
		Doc("get a user").
		Param(ws.PathParameter("user-id", "identifier of the user").DataType("string")).
		Writes(User{}))		
	...
	
	func (u UserResource) findUser(request *restful.Request, response *restful.Response) {
		id := request.PathParameter("user-id")
		...
	}
	
[Full API of a UserResource](https://github.com/emicklei/go-restful/tree/master/examples/restful-user-resource.go) 
		
### Features

- Routes for request -> function mapping with path parameter (e.g. {id}) support
- Configurable router:
	- Routing algorithm after [JSR311](http://jsr311.java.net/nonav/releases/1.1/spec/spec.html) that is implemented using (but doest **not** accept) regular expressions (See RouterJSR311 which is used by default)
	- Fast routing algorithm that only allows static elements, regular expressions and dynamic parameters in the URL path (e.g. /meetings/{id} or /static/{subpath:*}, See CurlyRouter)
- Request API for reading structs from JSON/XML and accesing parameters (path,query,header)
- Response API for writing structs to JSON/XML and setting headers
- Filters for intercepting the request &rightarrow; response flow	 on Service or Route level
- Request-scoped variables using attributes
- Containers for WebServices on different HTTP endpoints
- Content encoding (gzip,deflate) of responses
- Automatic responses on OPTIONS (using a filter)
- Automatic CORS request handling (using a filter)
- API declaration for Swagger UI (see swagger package)
- Panic recovery to produce HTTP 500, customizable using RecoverHandler(...)
	
### Resources

- [Documentation on godoc.org](http://godoc.org/github.com/emicklei/go-restful)
- [Code examples](https://github.com/emicklei/go-restful/tree/master/examples)
- [Example posted on blog](http://ernestmicklei.com/2012/11/24/go-restful-first-working-example/)
- [Design explained on blog](http://ernestmicklei.com/2012/11/11/go-restful-api-design/)
- [Sourcegraph](https://sourcegraph.com/github.com/emicklei/go-restful)
- [Showcase: Mora - MongoDB REST Api server](https://github.com/emicklei/mora)

[![Build Status](https://drone.io/github.com/emicklei/go-restful/status.png)](https://drone.io/github.com/emicklei/go-restful/latest)

[![library users](https://sourcegraph.com/api/repos/github.com/emicklei/go-restful/badges/library-users.png)](https://sourcegraph.com/github.com/emicklei/go-restful) [![authors](https://sourcegraph.com/api/repos/github.com/emicklei/go-restful/badges/authors.png)](https://sourcegraph.com/github.com/emicklei/go-restful) [![xrefs](https://sourcegraph.com/api/repos/github.com/emicklei/go-restful/badges/xrefs.png)](https://sourcegraph.com/github.com/emicklei/go-restful)

(c) 2013, http://ernestmicklei.com. MIT License