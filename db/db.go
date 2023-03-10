package db

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/google/uuid"
)

type Database struct {
	client    *dynamodb.DynamoDB
	tablename string
}

func (db Database) CreateMovie(movie Movie) (Movie, error) {
	movie.Id = uuid.New().String()
	entityParsed, err := dynamodbattribute.MarshalMap(movie)
	if err != nil {
		return Movie{}, err
	}

	input := &dynamodb.PutItemInput{
		Item:      entityParsed,
		TableName: aws.String(db.tablename),
	}

	_, err = db.client.PutItem(input)
	if err != nil {
		return Movie{}, err
	}

	return movie, nil
}

func (db Database) GetMovies() ([]Movie, error) {
	movies := []Movie{}
	filt := expression.Name("Id").AttributeNotExists()
	proj := expression.NamesList(
		expression.Name("id"),
		expression.Name("name"),
	)
	expr, err := expression.NewBuilder().WithFilter(filt).WithProjection(proj).Build()
	if err != nil {
		return []Movie{}, err
	}
	params := &dynamodb.ScanInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		ProjectionExpression:      expr.Projection(),
		TableName:                 aws.String(db.tablename),
	}
	result, err := db.client.Scan(params)

	if err != nil {

		return []Movie{}, err
	}

	for _, item := range result.Items {
		var movie Movie
		err = dynamodbattribute.UnmarshalMap(item, &movie)
		movies = append(movies, movie)

	}

	return movies, nil
}

func (db Database) GetMovie(id string) (Movie, error) {
	result, err := db.client.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(db.tablename),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(id),
			},
		},
	})
	if err != nil {
		return Movie{}, err
	}
	if result.Item == nil {
		msg := fmt.Sprintf("Movie with id [ %s ] not found", id)
		return Movie{}, errors.New(msg)
	}
	var movie Movie
	err = dynamodbattribute.UnmarshalMap(result.Item, &movie)
	if err != nil {
		return Movie{}, err
	}

	return movie, nil
}

func (db Database) UpdateMovie(movie Movie) (Movie, error) {
	entityParsed, err := dynamodbattribute.MarshalMap(movie)
	if err != nil {
		return Movie{}, err
	}

	input := &dynamodb.PutItemInput{
		Item:      entityParsed,
		TableName: aws.String(db.tablename),
	}

	_, err = db.client.PutItem(input)
	if err != nil {
		return Movie{}, err
	}

	return movie, nil
}

func (db Database) DeleteMovie(id string) error {
	input := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(id),
			},
		},
		TableName: aws.String(db.tablename),
	}

	res, err := db.client.DeleteItem(input)
	if res == nil {
		return errors.New(fmt.Sprintf("No movie to delete: %s", err))
	}
	if err != nil {
		return errors.New(fmt.Sprintf("Got error calling DeleteItem: %s", err))
	}
	return nil
}

type Movie struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type MovieService interface {
	CreateMovie(m Movie) (Movie, error)
	GetMovies() ([]Movie, error)
	GetMovie(id string) (Movie, error)
	UpdateMovie(m Movie) (Movie, error)
	DeleteMovie(id string) error
}

func InitDatabase() MovieService {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	return &Database{
		client:    dynamodb.New(sess),
		tablename: "Movies",
	}
}
