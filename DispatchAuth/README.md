# MyLibrary

Go API Server
golib-api-server makes it easy to create a standards conforming API server. Currently only JSON-RPC over HTTP is supported, but the architecture was written with other formats in mind.

Why not use net/rpc/jsonrpc?
The error reporting in that library is poor, and it doesn't support any validation. Schema validation helps prevent bad input, and it also serves as API documentation. golib-api-server requires each RPC call to have a JSON schema for the request and the response, and this schema is strictly enforced. Validation failures are reported to the API client, sometimes with an excerpt highlighting the problem.

How to use
First, add the library to your project: dep ensure -add gerrit.ingrooves.com/golib-api-server.git The .git extension acts as a hint which version control system we use, and is needed for the time being when using dep.

Define a structure that binds the protocol with the transport and the application specific logic (in this example we will use JSON-RPC):

type jsonRPCServer struct {
	transport rpc.ServerTransport
	protocol  *jsonrpc.ServerProtocol
	logic     *logic
}
Add a way to start and stop the server:

func (s *jsonRPCServer) Start() {
	s.transport.Start(s.protocol)
}

func (s *jsonRPCServer) Stop() *failure.Reason {
	if s.logic != nil {
		s.logic.close()
		s.logic = nil
	}

	return s.transport.Stop()
}
Add a constructor and the definition of all the supported API methods:

func NewJSONRPCServer(t rpc.ServerTransport, schemasDirectory string) (Server, *failure.Reason) {
	p, failed := jsonrpc.NewServerProtocol(schemasDirectory)
	if failed != nil {
		return nil, failed
	}

	s := jsonRPCServer{
		transport: t,
		protocol:  p,
	}

	l := newLogic(db, clock)

	if failed := s.registerMethods(l); failed != nil {
		l.close()

		return nil, failed
	}

	return &s, nil
}

func (s *jsonRPCServer) registerMethods(l *logic) *failure.Reason {
	m := jsonrpc.Method{
		Name:                   "addDelivery",                  // The method name as it will appear in JSON-RPC calls
		Handler:                l.addDelivery,                  // The function to call
		RequestSchemaFileName:  "add_delivery_request.json",    // The filename of the JSON schema the request will be validated against
		ResponseSchemaFileName: "add_delivery_response.json",   // The filename of the JSON schema the response will be validated against
	}

	if failed := s.protocol.RegisterMethod(m); failed != nil {
		return failed
	}
    
    ...etc
}
In main() bind the JSON-RPC protocol handler with a HTTP transport, and start listening for requests:

httpServer := http.NewServerTransport(apiConfig.API.Port, apiConfig.API.MaxRequestBodyLength, apiConfig.API.AllowedOrigins)
	api, failed := api.NewJSONRPCServer(httpServer, apiConfig.API.Schemas)
	if failed != nil {
		// Log this
		return
	}

	api.Start()
	defer func() {
		fmt.Println("Stopping listening for API calls.")
		api.Stop()
	}()
Defining the logic for calls
Let's define what will happen when an API client calls the “addDelivery” method from the example above.

type AddDeliveryParams struct {
	Delivery delivery.Delivery `json:"delivery"`
}

type addDeliveryResult struct {
	DeliveryID string `json:"deliveryId"`
}

func (l *logic) addDelivery(params AddDeliveryParams) (addDeliveryResult, *failure.Reason) {
    // Do something
    
	if failed != nil {
		return addDeliveryResult{}, failed
	}

	return addDeliveryResult{DeliveryID: id}, nil
}
You‘ll notice the input and outputs are plain Go structs. Besides the naming annotations json:"name", the application logic doesn’t know what protocol the caller is using.

Consider keeping each call its own .go file like this example: , which also contains the input and output structures - like the example above.

Adding protocols
Adding transport
Powered by Gitiles
source
log
blame