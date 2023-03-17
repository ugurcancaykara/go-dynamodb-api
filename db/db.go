package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/google/uuid"
	instana "github.com/instana/go-sensor"
	"github.com/instana/go-sensor/instrumentation/instaawssdk"
)

type Database struct {
	client    *dynamodb.DynamoDB
	tablename string
}

func (db Database) CreateMovie(ctx context.Context, movie Movie) (Movie, error) {
	movie.Id = uuid.New().String()
	entityParsed, err := dynamodbattribute.MarshalMap(movie)
	if err != nil {
		return Movie{}, err
	}

	input := &dynamodb.PutItemInput{
		Item:      entityParsed,
		TableName: aws.String(db.tablename),
	}

	_, err = db.client.PutItemWithContext(ctx, input)
	if err != nil {
		return Movie{}, err
	}

	return movie, nil
}

func (db Database) GetMovies(ctx context.Context) ([]Movie, error) {
	movies := []Movie{}
	filt := expression.Name("Id").AttributeNotExists()
	proj := expression.NamesList(
		expression.Name("id"),
		expression.Name("name"),
		expression.Name("description"),
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

	result, err := db.client.ScanWithContext(ctx, params)

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

func (db Database) GetMovie(ctx context.Context, id string) (Movie, error) {
	result, err := db.client.GetItemWithContext(ctx, &dynamodb.GetItemInput{
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

func newDynamoDBRequest(db Database, entityParsed map[string]*dynamodb.AttributeValue) *request.Request {
	svc := db.client
	req, _ := svc.PutItemRequest(&dynamodb.PutItemInput{
		Item:      entityParsed,
		TableName: aws.String(db.tablename),
	})
	return req
}

func (db Database) UpdateMovie(movie Movie, ctx context.Context, sensor *instana.Sensor) (Movie, error) {
	entityParsed, err := dynamodbattribute.MarshalMap(movie)

	parentSp := sensor.Tracer().StartSpan("testing", opentracing.Tags{
		"dynamodb.op":     "get",
		"dynamodb.table":  "test-table",
		"dynamodb.region": "mock-region",
	})

	if err != nil {
		return Movie{}, err
	}

	input := &dynamodb.PutItemInput{
		Item:      entityParsed,
		TableName: aws.String(db.tablename),
	}
	req := newDynamoDBRequest(db, entityParsed)

	req.SetContext(instana.ContextWithSpan(req.Context(), parentSp))
	instaawssdk.StartDynamoDBSpan(req, sensor)
	//fmt.Println(req)
	sp, _ := instana.SpanFromContext(req.Context())

	_, err = db.client.PutItemWithContext(req.Context(), input)
	if err != nil {
		return Movie{}, err
	}

	defer sp.Finish()
	defer parentSp.Finish()
	instaawssdk.FinalizeDynamoDBSpan(req)
	return movie, nil
}

func (db Database) DeleteMovie(ctx context.Context, id string) error {
	input := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(id),
			},
		},
		TableName: aws.String(db.tablename),
	}

	res, err := db.client.DeleteItemWithContext(ctx, input)
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
	CreateMovie(ctx context.Context, m Movie) (Movie, error)
	GetMovies(ctx context.Context) ([]Movie, error)
	GetMovie(ctx context.Context, id string) (Movie, error)
	UpdateMovie(m Movie, ctx context.Context, sensor *instana.Sensor) (Movie, error)
	DeleteMovie(ctx context.Context, id string) error
}

func InitDatabase(sensor *instana.Sensor) MovieService {
	//sess := session.Must(session.NewSessionWithOptions(session.Options{
	//	SharedConfigState: session.SharedConfigEnable,
	//}))

	//sess := session.Must(session.NewSession(&aws.Config{
	//	Region: aws.String("eu-west-1"),
	//}))

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		// Note: Please use the appropriate region for testing.
		Config:  aws.Config{Region: aws.String("us-east-2")},
		Profile: "default",
	}))

	instaawssdk.InstrumentSession(sess, sensor)

	return &Database{
		client: dynamodb.New(sess),
		// Note: Please use the appropriate database for testing.
		tablename: "go-sensor-movies",
	}
}
