package controllers

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/tmc/langchaingo/llms/googleai"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	database "github.com/aKeles001/golang_movie/Server/magic_movies_server/database"
	models "github.com/aKeles001/golang_movie/Server/magic_movies_server/models"
	utils "github.com/aKeles001/golang_movie/Server/magic_movies_server/utils"
)

var validate = validator.New()

func GetMovies(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var movies []models.Movie
		movieCollection := database.OpenCollection("movies", client)
		cursor, err := movieCollection.Find(ctx, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching movies from database"})
		}
		defer cursor.Close(ctx)
		if err = cursor.All(ctx, &movies); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding movies from database"})
		}
		c.JSON(http.StatusOK, movies)
	}
}

func GetMovie(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		movieID := c.Param("imdb_id")
		if movieID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid movie ID"})
			return
		}
		var movieCollection *mongo.Collection = database.OpenCollection("movies", client)
		var movie models.Movie
		if err := movieCollection.FindOne(ctx, bson.M{"imdb_id": movieID}).Decode(&movie); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching movie from database"})
			return
		}
		c.JSON(http.StatusOK, movie)
	}
}

func AddMovie(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var validate = validator.New()
		var movie models.Movie
		if err := c.ShouldBindJSON(&movie); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON provided"})
			return
		}
		if err := validate.Struct(movie); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
			return
		}
		var movieCollection *mongo.Collection = database.OpenCollection("movies", client)
		result, err := movieCollection.InsertOne(ctx, movie)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error inserting movie into database"})
			return
		}
		c.JSON(http.StatusCreated, result)
	}
}

func AdminReviewUpdate(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {

		role, err := utils.GetRoleFromContext(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		if role != "ADMIN" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			return
		}

		movieId := c.Param("imdb_id")
		if movieId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid movie ID"})
			return
		}
		var req struct {
			AdminReview string `json:"admin_review"`
		}
		var response struct {
			RankingName string `json:"ranking_name"`
			AdminReview string `json:"admin_review"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		sentiment, rankValue, err := GetReviewRanking(req.AdminReview, client)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error analyzing review", "details": err.Error()})
			return
		}
		filter := bson.M{"imdb_id": movieId}
		update := bson.M{
			"$set": bson.M{
				"admin_review": req.AdminReview,
				"ranking": bson.M{
					"ranking_name":  sentiment,
					"ranking_value": rankValue,
				},
			},
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var movieCollection *mongo.Collection = database.OpenCollection("movies", client)
		result, err := movieCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating movie review"})
			return
		}
		if result.MatchedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Movie not found"})
			return
		}

		response.RankingName = sentiment
		response.AdminReview = req.AdminReview
		c.JSON(http.StatusOK, response)
	}
}

func GetReviewRanking(admin_review string, client *mongo.Client) (string, int, error) {
	rankings, err := GetRankings(client)
	if err != nil {
		return "", 0, err
	}
	sentimentDelimited := ""
	for _, ranking := range rankings {
		if ranking.RankingValue != 999 {
			sentimentDelimited = sentimentDelimited + ranking.RankingName + ","
		}
	}

	sentimentDelimited = strings.Trim(sentimentDelimited, ",")
	llm, err := getLLM()
	if err != nil {
		return "", 0, err
	}
	promptTemplate := os.Getenv("BASE_PROMT_TEMPLATE")
	prompt := strings.Replace(promptTemplate, "{rankings}", sentimentDelimited, 1)
	response, err := llm.Call(context.Background(), prompt+" "+admin_review)
	if err != nil {
		log.Printf("Error from Gemini API: %v", err)
		return "", 0, err
	}
	rankValue := 0
	for _, ranking := range rankings {
		if ranking.RankingName == response {
			rankValue = ranking.RankingValue
			break
		}
	}
	return response, rankValue, nil
}

func GetRankings(client *mongo.Client) ([]models.Ranking, error) {
	var rankings []models.Ranking

	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var rankingsCollection *mongo.Collection = database.OpenCollection("rankings", client)
	cursor, err := rankingsCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &rankings); err != nil {
		return nil, err
	}
	return rankings, nil
}

var (
	llm    *googleai.GoogleAI
	llmErr error
	once   sync.Once
)

func getLLM() (*googleai.GoogleAI, error) {
	once.Do(func() {
		GoogleAiApiKey := os.Getenv("GOOGLE_API_KEY")
		if GoogleAiApiKey == "" {
			llmErr = errors.New("could not find GOOGLE_API_KEY in environment variables")
			return
		}
		var err error
		llm, err = googleai.New(context.Background(), googleai.WithAPIKey(GoogleAiApiKey))
		if err != nil {
			llmErr = err
		}
	})
	return llm, llmErr
}

func GetReqomendedMovies(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		userId, err := utils.GetUserIdFromContext(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		favoriteGenres, err := GetUsersFavoriteGenres(userId, client)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching user's favorite genres"})
			return
		}
		var recommendedLimitVal int64 = 5

		recommendedLimitStr := os.Getenv("RECOMMENDED_LIMIT_VALUE")
		if recommendedLimitStr != "" {
			recommendedLimitVal, _ = strconv.ParseInt(recommendedLimitStr, 10, 64)
		}
		findOptions := options.Find()
		findOptions.SetSort(bson.D{{Key: "ranking.ranking_value", Value: 1}})
		findOptions.SetLimit(recommendedLimitVal)

		filter := bson.D{
			{Key: "genre.genre_name", Value: bson.D{
				{Key: "$in", Value: favoriteGenres},
			}},
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var movieCollection *mongo.Collection = database.OpenCollection("movies", client)
		cursor, err := movieCollection.Find(ctx, filter, findOptions)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching recommended movies from database"})
			return
		}
		defer cursor.Close(ctx)

		var recommendedMovies []models.Movie
		if err = cursor.All(ctx, &recommendedMovies); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding recommended movies from database"})
			return
		}
		c.JSON(http.StatusOK, recommendedMovies)
	}
}

func GetUsersFavoriteGenres(userId string, client *mongo.Client) ([]string, error) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	filter := bson.M{"user_id": userId}

	projection := bson.M{
		"favorite_genres.genre_name": 1,
		"_id":                        0,
	}

	opts := options.FindOne().SetProjection(projection)
	var result bson.M

	var userCollection *mongo.Collection = database.OpenCollection("users", client)
	err := userCollection.FindOne(ctx, filter, opts).Decode(&result)
	if err == mongo.ErrNoDocuments {
		return []string{}, nil
	}
	favGenresArray, ok := result["favorite_genres"].(bson.A)
	if !ok {
		return []string{}, errors.New("Failed to parse favorite genres")
	}
	var genreNames []string
	for _, genre := range favGenresArray {
		if genreMap, ok := genre.(bson.D); ok {
			for _, item := range genreMap {
				if item.Key == "genre_name" {
					if name, ok := item.Value.(string); ok {
						genreNames = append(genreNames, name)

					}
				}
			}
		}
	}
	return genreNames, nil
}
