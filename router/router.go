package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instagin"
	dynamodb "go-crud-api/db"
)

var iSensor = instana.NewSensorWithTracer(instana.NewTracerWithOptions(&instana.Options{
	Service:           "test-sensor-3",
	LogLevel:          instana.Debug,
	EnableAutoProfile: true,
},
))

var db = dynamodb.InitDatabase(iSensor)

func InitRouter() *gin.Engine {

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

	c := ctx.Request.Context()

	res, err := db.GetMovies(c)
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
	res, err := db.GetMovie(ctx.Request.Context(), id)
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
	res, err := db.CreateMovie(ctx.Request.Context(), movie)
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
	res, err := db.GetMovie(ctx.Request.Context(), id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}
	res.Name = movie.Name
	res.Description = movie.Description

	res, err = db.UpdateMovie(res, ctx, iSensor)

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
	err := db.DeleteMovie(ctx.Request.Context(), id)
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
