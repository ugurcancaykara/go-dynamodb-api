package router

import (
	"github.com/gin-gonic/gin"
	dynamodb "go-crud-api/db"
	"net/http"
)

var db = dynamodb.InitDatabase()

func InitRouter() *gin.Engine {
	r := gin.Default()
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
	res, err = db.UpdateMovie(res)

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
