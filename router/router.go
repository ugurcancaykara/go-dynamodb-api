package router

import (
	"github.com/gin-gonic/gin"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instagin"
	dynamodb "go-crud-api/db"
	"net/http"
)

var recorder = instana.NewTestRecorder()
var iSensor = instana.NewSensorWithTracer(
	instana.NewTracerWithEverything(&instana.Options{}, recorder),
)

// var iSensor = instana.NewSensor("my-first-sensor")
var db = dynamodb.InitDatabase(iSensor)

//var iSensor *instana.Sensor

func InitRouter() *gin.Engine {

	// If we just use initsensor and not tracer, we won't be able to trace requests.
	//instana.InitSensor(&instana.Options{
	//	Service:           "my-movie-app",
	//	LogLevel:          instana.Debug,
	//	EnableAutoProfile: true,
	//})
	//instana.InitSensor(instana.DefaultOptions())

	//iSensor = instana.NewSensor(
	//	"my-movie-app-tracing",
	//)
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

	r := gin.Default()
	instagin.AddMiddleware(iSensor, r)
	r.GET("/movies", getMovies)
	r.GET("/movies/:id", getMovie)
	r.POST("/movies", postMovie)
	r.PUT("/movies/:id", putMovie)
	r.DELETE("/movies/:id", deleteMovie)
	return r
}

func getMovies(ctx *gin.Context) {

	res, err := db.GetMovies()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"movies": res,
	})
}

func getMovie(ctx *gin.Context) {
	id := ctx.Param("id")
	res, err := db.GetMovie(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
	}
	ctx.JSON(http.StatusOK, gin.H{
		"movie": res,
	})
}

func postMovie(ctx *gin.Context) {
	var movie dynamodb.Movie
	err := ctx.ShouldBind(&movie)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	res, err := db.CreateMovie(movie)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	ctx.IndentedJSON(http.StatusOK, gin.H{
		"movies": res,
	})
}

func putMovie(ctx *gin.Context) {
	var movie dynamodb.Movie
	err := ctx.ShouldBind(&movie)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	id := ctx.Param("id")
	res, err := db.GetMovie(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}
	res.Name = movie.Name
	res.Description = movie.Description
	res, err = db.UpdateMovie(res, ctx, iSensor, recorder)

	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}
	ctx.IndentedJSON(http.StatusOK, gin.H{
		"movie": res,
	})
}

func deleteMovie(ctx *gin.Context) {
	id := ctx.Param("id")
	err := db.DeleteMovie(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"movie": "Movie deleted successfully",
	})
}
