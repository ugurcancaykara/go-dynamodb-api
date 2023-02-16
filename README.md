# go-dynamodb-api

Basic crud api using gin web framework and dynamodb as database for to test Instana go sensor


For http applications:
    // If we just use initsensor and not tracer, we won't be able to trace requests.
    //instana.InitSensor(&instana.Options{
    //	Service:           "my-movie-app",
    //	LogLevel:          instana.Debug,
    //	EnableAutoProfile: true,
    //})
    //instana.InitSensor(instana.DefaultOptions())
    
    iSensor = instana.NewSensor(
        "my-movie-app-tracing",
    )
    // Using Newsensor with options(in most cases, leaving the default configuration options in place will be enough)
    // https://pkg.go.dev/github.com/instana/go-sensor#TracerOptions
    //instana.NewSensorWithTracer(instana.NewTracerWithOptions(&instana.Options{
    //	Service:           "my-movie-app-tracing",
    //	EnableAutoProfile: true,
    //},
    //))
    
    // collect extra HTTP headers https://www.ibm.com/docs/en/instana-observability/current?topic=go-collector-common-operations#how-to-collect-extra-http-headers
    //instana.NewSensorWithTracer(instana.NewTracerWithOptions(&instana.Options{
    //	Service:           "my-movie-app-tracing",
    //	EnableAutoProfile: true,
    //	Tracer: instana.TracerOptions{
    //		CollectableHTTPHeaders: []string{"x-request-id","x-loadtest-id"},
    //	},
    //},
    //))

For grpc applications:
- https://pkg.go.dev/github.com/instana/go-sensor/instrumentation/instagrpc#section-readme